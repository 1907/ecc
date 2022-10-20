package importing

import (
	"Ecc/configs"
	"Ecc/pkg/files"
	"Ecc/pkg/mongo"
	"Ecc/pkg/mysql"
	"Ecc/tools"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	ProcessBarWidth = 48
	Exts            = ".xlsx,.xls,.csv"
)

var (
	rWait   = true
	wWait   = true
	rDone   = make(chan struct{})
	rCrash  = make(chan struct{})
	wDone   = make(chan struct{})
	wCrash  = make(chan struct{})
	once    = &sync.Once{}
	wg      = &sync.WaitGroup{}
	pb      = mpb.New(mpb.WithWaitGroup(wg), mpb.WithWidth(ProcessBarWidth))
	shuttle = Shuttle{}
)

type File struct {
	FileName string
	FilePath string
}

func Load(cf string) {
	var err error
	viper.SetConfigFile(cf)
	if err = viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("fatal error config file: %s \n", err))
	}
	if err = viper.Unmarshal(&configs.C); err != nil {
		log.Fatal(fmt.Errorf("unmarshal conf failed, err:%s \n", err))
	}
	if err = mongo.Conn(configs.C.Mongo.DNS, configs.C.Mongo.Db); err != nil {
		log.Fatal(color.RedString("%s:\n%v", "mongo connect err", err))
	}
	if mongo.CheckCollection(configs.C.Mongo.Collection) { // del mongo collection if exists
		if err = mongo.DelCollection(configs.C.Mongo.Collection); err != nil {
			log.Fatal(color.RedString("%s:\n%v", "mongo del collection err", err))
		}
	}
	if err = mongo.CreateCollection(configs.C.Mongo.Collection); err != nil {
		log.Fatal(color.RedString("%s:\n%v", "mongo create collection err", err))
	}
}

func Handle(dir string) {
	var err error
	var f []os.FileInfo
	var data = &sync.Map{}

	if f, err = files.ReadDir(dir); err != nil {
		abort("-> Failure: " + err.Error())
		return
	}

	read(f, dir, data)
	for rWait {
		select {
		case <-rCrash:
			abort("-> Failure")
			return
		case <-rDone:
			rWait = false
		}
	}

	write2mongo(data)
	for wWait {
		select {
		case <-wCrash:
			abort("-> Failure")
			return
		case <-wDone:
			wWait = false
		}
	}

	pb.Wait()

	tools.Yellow("-> Whether to sync data to mysql? (y/n)")
	if !tools.Scan("aborted") {
		return
	} else {
		tools.Yellow("-> Syncing data to mysql...")
		if err = write2mysql(); err != nil {
			tools.Red("-> Failure：" + err.Error())
		} else {
			tools.Green("-> Success.")
		}
	}
}

func read(fs []os.FileInfo, dir string, data *sync.Map) {
	for _, file := range fs {
		fileName := file.Name()
		_ext := filepath.Ext(fileName)
		if Include(strings.Split(Exts, ","), _ext) {
			wg.Add(1)
			inCh := make(chan File)
			go func() {
				defer wg.Done()
				select {
				case <-rCrash:
					return // exit
				case f := <-inCh:
					e, preData := ReadExcel(f.FilePath, f.FileName, pb)
					if e != nil {
						tools.Red("%v", e)
						once.Do(func() {
							close(rCrash)
						})
						return
					}
					data.Store(f.FileName, preData)
				}
			}()
			go func() {
				inCh <- File{
					FileName: fileName,
					FilePath: dir + string(os.PathSeparator) + fileName,
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(rDone)
	}()
}

func write2mongo(data *sync.Map) {
	collection := mongo.GetCollection(configs.C.Mongo.Collection)
	data.Range(func(key, value interface{}) bool {
		if v, ok := value.(*ExcelPre); ok {
			wg.Add(1)
			inCh := make(chan []bson.M)
			go func() {
				defer wg.Done()
				select {
				case <-wCrash:
					return // exit
				case rows := <-inCh:
					e := Write2Mongo(rows, collection, v, &shuttle)
					if e != nil {
						tools.Red("%v", e)
						once.Do(func() {
							close(wCrash)
						})
						return
					}
				}
			}()
			go func() {
				inCh <- PreWrite(v)
			}()
		}
		return true
	})

	go func() {
		wg.Wait()
		close(wDone)
	}()
}

func write2mysql() error {
	if err := mysql.Conn(configs.C.Mysql.Dns); err != nil {
		return err
	}

	d, err := mongo.GetCollectionAllData(configs.C.Mongo.Collection)
	if err != nil {
		return err
	}

	err = Write2Mysql(d)
	return err
}

func abort(e string) {
	if rWait == false {
		records := make([]primitive.ObjectID, 0)
		for _, id := range shuttle.Mid {
			records = append(records, mongo.GetObjectId(id))
		}

		_ = mongo.DeleteMany(
			configs.C.Mongo.Collection,
			bson.M{"_id": map[string][]primitive.ObjectID{"$in": records}})
	}

	tools.Red(e)
}
