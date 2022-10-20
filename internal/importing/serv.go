package importing

import (
	"Ecc/pkg/mongo"
	"Ecc/pkg/mysql"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDb "go.mongodb.org/mongo-driver/mongo"
	"regexp"
	"sync"
	"time"
)

type ExcelPre struct {
	FileName    string
	Data        [][]string
	Fields      []string
	Prefixes    string
	ProgressBar *mpb.Bar
}

type Shuttle struct {
	Hid []string
	Mid []string
	mu  sync.Mutex
}

func (s *Shuttle) Append(t string, str string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch t {
	case "h":
		s.Hid = append(s.Hid, str)
	case "m":
		s.Mid = append(s.Mid, str)
	}
}

func ReadExcel(filePath, fileName string, pb *mpb.Progress) (err error, pre *ExcelPre) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err, nil
	}

	defer func() {
		if _e := f.Close(); _e != nil {
			fmt.Printf("%s: %v.\n\n", filePath, _e)
		}
	}()

	// Get the first sheet.
	firstSheet := f.WorkBook.Sheets.Sheet[0].Name
	rows, err := f.GetRows(firstSheet)
	lRows := len(rows)

	if lRows < 2 {
		lRows = 2
	}

	rb := ReadBar(lRows, filePath, pb)
	wb := WriteBar(lRows-2, filePath, rb, pb)

	// The first line is the field name.
	var fields []string

	// The data of the file.
	var data [][]string

	InCr := func(start time.Time) {
		rb.Increment()
		rb.DecoratorEwmaUpdate(time.Since(start))
	}

	for i := 0; i < lRows; i++ {
		InCr(time.Now())
		if i == 0 {
			fields = rows[i]
			for index, field := range fields {
				if isChinese := regexp.MustCompile("[\u4e00-\u9fa5]"); isChinese.MatchString(field) || field == "" {
					err = errors.New(fmt.Sprintf("%s: line 【A%d】 field 【%s】 \n", filePath, index, field) + "The first line of the file is not a valid attribute name.")
					return err, nil
				}
			}
			continue
		}

		if i == 1 {
			continue
		}

		data = append(data, rows[i])
	}

	return nil, &ExcelPre{
		FileName:    fileName,
		Data:        data,
		Fields:      fields,
		Prefixes:    Prefix(fileName),
		ProgressBar: wb,
	}
}

func PreWrite(v *ExcelPre) []bson.M {
	return CreateBm(
		v.Data,
		v.Fields,
		v.Prefixes,
		NewRules("hid", "splitting"),
	)
}

func Write2Mongo(rows []bson.M, collection *mongoDb.Collection, v *ExcelPre, s *Shuttle) error {
	v.ProgressBar.SetCurrent(0)
	incr := func(t time.Time, b *mpb.Bar, n int64) {
		b.IncrInt64(n)
		b.DecoratorEwmaUpdate(time.Since(t))
	}
	for _, row := range rows {
		start := time.Now()
		key := v.Prefixes + "@@" + row["_hid"].(string)

		s.mu.Lock()
		if Include(s.Hid, key) {
			s.mu.Unlock()
			incr(start, v.ProgressBar, 1)
			continue
		} else {
			s.Hid = append(s.Hid, key)
			s.mu.Unlock()
		}

		var err error
		var id primitive.ObjectID
		if id, err = mongo.CreateDocs(collection, row); err != nil {
			return errors.New(fmt.Sprintf("%s:\n%v", "mongo create docs err", err))
		}

		s.Append("m", id.Hex())
		incr(start, v.ProgressBar, 1)
	}

	return nil
}

func Write2Mysql(data []bson.M) error {
	var err error
	var m = make(map[string][]bson.M)
	var t = time.Now().Format("2006_01_02")

	for _, item := range data {
		for attr, v := range item {
			if attr == "type" {
				vs := v.(string)
				if _, ok := m[vs]; ok {
					m[vs] = append(m[vs], item)
				} else {
					m[vs] = make([]bson.M, 0)
				}
			}
		}
	}

	for table, d := range m {
		var fields []string
		table = table + "_" + t
		if len(d) > 0 {
			d0 := d[0]
			for f := range d0 {
				if f != "type" && f != "_id" && f != "_hid" {
					fields = append(fields, f)
				}
			}
		}
		if err = mysql.CreateTable(table, fields); err != nil {
			return err
		}
		if err = mysql.InsertData(table, fields, ExecRules(d, NewRules("revHid", "revSplitting"))); err != nil {
			return err
		}
	}

	return nil
}

func ReadBar(total int, name string, pb *mpb.Progress) *mpb.Bar {
	return pb.AddBar(int64(total),
		mpb.PrependDecorators(
			decor.OnComplete(decor.Name(color.YellowString("reading"), decor.WCSyncSpaceR), color.YellowString("waiting")),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth, decor.WCSyncSpaceR),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%.2f:", decor.WCSyncSpaceR),
			decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth),
			decor.Name(": "+name),
		),
	)
}

func WriteBar(total int, name string, beforeBar *mpb.Bar, pb *mpb.Progress) *mpb.Bar {
	return pb.AddBar(int64(total),
		mpb.BarQueueAfter(beforeBar, false),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.OnComplete(decor.Name(color.YellowString("writing"), decor.WCSyncSpaceR), color.GreenString("done")),
			decor.OnComplete(decor.CountersNoUnit("%d / %d", decor.WCSyncSpaceR), ""),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.NewPercentage("%.2f:", decor.WCSyncSpaceR), ""),
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 0, decor.WCSyncWidth), ""),
			decor.OnComplete(decor.Name(": "+name), name),
		),
	)
}
