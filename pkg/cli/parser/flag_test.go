package parser

import (
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFlag(t *testing.T) {
	Convey("test flag", t, func() {
		flagSet := NewFlagSet("test flag")
		i := flagSet.Int("i", 10, "int flag")
		f := flagSet.Float64("f", 11.11, "float flag")
		s := flagSet.String("s", "hello world", "string flag")
		d := flagSet.Duration("d", time.Duration(30)*time.Second, "string flag")
		b := flagSet.Bool("b", false, "bool flag")
		vi := flagSet.IntSlice("vi", []int{1, 2, 3, 4, 5}, "int slice flag")
		ip := flagSet.IP("ip", nil, "ip flag")
		t := flagSet.Time("time", time.Now(), "time flag")

		Convey("check default Value", func() {
			So(*i, ShouldEqual, 10)
			So(*f, ShouldAlmostEqual, 11.11)
			So(*s, ShouldEqual, "hello world")
			So(*d, ShouldEqual, time.Duration(30)*time.Second)
			So(*b, ShouldBeFalse)
			So(*vi, ShouldResemble, []int{1, 2, 3, 4, 5})
		})

		Convey("parse case success", func() {
			err := flagSet.Parse(strings.Split("-b -i 100 -f=12.12 --s golang --d=20s -vi 6,7,8,9 -ip 61.48.69.8 -time 2019-11-27", " "))
			So(err, ShouldBeNil)
			So(*i, ShouldEqual, 100)
			So(*f, ShouldAlmostEqual, 12.12)
			So(*s, ShouldEqual, "golang")
			So(*d, ShouldEqual, time.Duration(20)*time.Second)
			So(*b, ShouldBeTrue)
			So(*vi, ShouldResemble, []int{6, 7, 8, 9})
			So((*ip).String(), ShouldEqual, "61.48.69.8")
			So((*t).String()[0:10], ShouldEqual, "2019-11-27")
		})

		Convey("parse case unexpected Value", func() {
			// -i 后面期望之后一个参数，100 会被当做 -i 的参数，101 会被当成位置参数，继续解析
			err := flagSet.Parse(strings.Split("-b -i 100 101 -f 12.12 -s golang -d 20s", " "))
			So(err, ShouldBeNil)
			So(*b, ShouldBeTrue)
			So(*i, ShouldEqual, 100)
			So(*f, ShouldEqual, 12.12)
			So(*s, ShouldEqual, "golang")
			So(*d, ShouldEqual, 20*time.Second)
			So(flagSet.NFlag(), ShouldEqual, 7)
			So(flagSet.NArg(), ShouldEqual, 1)
			So(flagSet.Args(), ShouldResemble, []string{
				"101",
			})
		})

		Convey("set case", func() {
			err := flagSet.Set("i", "120")
			So(err, ShouldBeNil)
			So(*i, ShouldEqual, 120)
		})

		Convey("visit case", func() {
			// 遍历所有设置过得选项
			flagSet.Visit(func(f *Flag) {
				fmt.Println(f.Name)
			})

			// 遍历所有选项
			flagSet.VisitAll(func(f *Flag) {
				fmt.Println(f.Name)
			})
		})

		Convey("lookup case", func() {
			f := flagSet.Lookup("i")
			So(f.Name, ShouldEqual, "i")
			So(f.DefValue, ShouldEqual, "10")
			So(f.Usage, ShouldEqual, "int flag")
			So(f.Value.String(), ShouldEqual, "10")
		})

		Convey("print defaults", func() {
			flagSet.PrintDefaults()
		})
	})
}

func TestFlagParse(t *testing.T) {
	Convey("test case1", t, func() {
		flagSet := NewFlagSet("test flag")
		flagSet.AddFlag("int-option", "usage", Shorthand("i"), Type("int"), Required(), DefaultValue("10"))
		flagSet.AddFlag("str-option", "usage", Shorthand("s"), Required())
		flagSet.AddFlag("key", "usage", Shorthand("k"), Type("float64"), Required())
		flagSet.AddFlag("all", "usage", Shorthand("a"), Type("bool"), Required())
		flagSet.AddFlag("user", "usage", Shorthand("u"), Type("bool"), Required())
		flagSet.AddFlag("password", "usage", Shorthand("p"), Type("string"), DefaultValue("654321"))
		flagSet.AddFlag("vs", "usage", Shorthand("v"), Type("[]string"), DefaultValue("dog,cat"))
		flagSet.AddPosFlag("pos1", "usage", Type("string"))
		flagSet.AddPosFlag("pos2", "usage", Type("string"))

		Convey("check default value", func() {
			So(flagSet.GetInt("int-option"), ShouldEqual, 10)
			So(flagSet.GetString("str-option"), ShouldEqual, "")
			So(flagSet.GetFloat64("key"), ShouldEqual, 0.0)
			So(flagSet.GetBool("all"), ShouldBeFalse)
			So(flagSet.GetBool("user"), ShouldBeFalse)
			So(flagSet.GetString("password"), ShouldEqual, "654321")
			So(flagSet.GetStringSlice("vs"), ShouldResemble, []string{"dog", "cat"})
			So(flagSet.GetString("pos1"), ShouldEqual, "")
			So(flagSet.GetString("pos2"), ShouldEqual, "")
		})

		Convey("parse case 1", func() {
			err := flagSet.Parse([]string{
				"val1",
				"--int-option=123",
				"--str-option", "apple,banana,orange",
				"-k", "3.14",
				"-au",
				"-p123456",
				"val2",
				"-vs", "one,two,three",
			})
			So(err, ShouldBeNil)

			So(flagSet.GetInt("int-option"), ShouldEqual, 123)
			So(flagSet.GetString("str-option"), ShouldEqual, "apple,banana,orange")
			So(flagSet.GetStringSlice("str-option"), ShouldResemble, []string{
				"apple", "banana", "orange",
			})
			So(flagSet.GetFloat64("key"), ShouldAlmostEqual, 3.14)
			So(flagSet.GetBool("all"), ShouldBeTrue)
			So(flagSet.GetBool("user"), ShouldBeTrue)
			So(flagSet.GetString("password"), ShouldEqual, "123456")
			So(flagSet.GetStringSlice("vs"), ShouldResemble, []string{"one", "two", "three"})
			So(flagSet.GetString("pos1"), ShouldEqual, "val1")
			So(flagSet.GetString("pos2"), ShouldEqual, "val2")

			So(flagSet.Args(), ShouldResemble, []string{
				"val1", "val2",
			})
			So(flagSet.Arg(0), ShouldEqual, "val1")
			So(flagSet.Arg(1), ShouldEqual, "val2")
		})

		Convey("parse case 2", func() {
			err := flagSet.Parse(strings.Split("--str-option=1,2,3,4 -key=3 -au", " "))
			So(err, ShouldBeNil)
			So(flagSet.GetIntSlice("str-option"), ShouldResemble, []int{1, 2, 3, 4})
			So(flagSet.GetString("a"), ShouldEqual, "true")
			So(flagSet.GetString("u"), ShouldEqual, "true")
		})
	})

	Convey("test case2", t, func() {
		flagSet := NewFlagSet("test flag")
		version := flagSet.Bool("v", false, "print current version")
		configfile := flagSet.String("c", "configs/monitor.json", "config file path")
		So(flagSet.Parse(strings.Split("--v", " ")), ShouldBeNil)
		So(*version, ShouldBeTrue)
		So(*configfile, ShouldEqual, "configs/monitor.json")
	})
}

func TestFlagAddFlags(t *testing.T) {
	Convey("test add flags", t, func() {
		flagSet := NewFlagSet("test flag")
		type MySubFlags struct {
			F1 int    `flag:"--f1; default:20; usage:f1 flag"`
			F2 string `flag:"--f2; default:hatlonely; usage:f2 flag"`
		}
		type MyFlags struct {
			IntOption int         `flag:"--int-option, -i; default:10; required; usage:int flag"`
			StrOption string      `flag:"--str-option, -s; required; usage:str flag"`
			Key       float64     `flag:"--key, -k; required; usage:float flag"`
			All       bool        `flag:"--all, -a; required; usage:bool flag"`
			User      bool        `flag:"--user, -u; required; usage:user flag"`
			Password  string      `flag:"--password, -p; default:654321; usage:password flag"`
			Vs        []string    `flag:"--vs, -v; default:dog,cat; usage:vs flag"`
			Pos1      string      `flag:"pos1; usage:string pos flag"`
			Pos2      string      `flag:"pos2; usage:int pos flag"`
			Sub       *MySubFlags `flag:"sub"`
		}
		f := &MyFlags{}
		So(flagSet.Bind(f), ShouldBeNil)

		Convey("check default value", func() {
			So(f.IntOption, ShouldEqual, 10)
			So(f.StrOption, ShouldEqual, "")
			So(f.Key, ShouldEqual, 0.0)
			So(f.All, ShouldBeFalse)
			So(f.User, ShouldBeFalse)
			So(f.Password, ShouldEqual, "654321")
			So(f.Vs, ShouldResemble, []string{"dog", "cat"})
			So(f.Pos1, ShouldEqual, "")
			So(f.Pos2, ShouldEqual, "")
			So(f.Sub.F1, ShouldEqual, 20)
		})

		Convey("parse case 1", func() {
			err := flagSet.Parse([]string{
				"val1",
				"--int-option=123",
				"--str-option", "apple,banana,orange",
				"-k", "3.14",
				"-au",
				"-p123456",
				"val2",
				"-vs", "one,two,three",
				"--sub-f1", "120",
			})
			So(err, ShouldBeNil)

			So(f.IntOption, ShouldEqual, 123)
			So(f.StrOption, ShouldEqual, "apple,banana,orange")
			So(flagSet.GetStringSlice("str-option"), ShouldResemble, []string{
				"apple", "banana", "orange",
			})
			So(f.Key, ShouldAlmostEqual, 3.14)
			So(f.All, ShouldBeTrue)
			So(f.User, ShouldBeTrue)
			So(f.Password, ShouldEqual, "123456")
			So(f.Vs, ShouldResemble, []string{"one", "two", "three"})
			So(f.Pos1, ShouldEqual, "val1")
			So(f.Pos2, ShouldEqual, "val2")

			So(flagSet.Args(), ShouldResemble, []string{
				"val1", "val2",
			})
			So(flagSet.Arg(0), ShouldEqual, "val1")
			So(flagSet.Arg(1), ShouldEqual, "val2")
			So(f.Sub.F1, ShouldEqual, 120)

			mf := &MyFlags{}
			So(flagSet.Unmarshal(mf), ShouldBeNil)
			So(mf.IntOption, ShouldEqual, 123)
			So(mf.StrOption, ShouldEqual, "apple,banana,orange")
			So(mf.Key, ShouldAlmostEqual, 3.14)
			So(mf.All, ShouldBeTrue)
			So(mf.User, ShouldBeTrue)
			So(mf.Password, ShouldEqual, "123456")
			So(mf.Vs, ShouldResemble, []string{"one", "two", "three"})
			So(mf.Pos1, ShouldEqual, "val1")
			So(mf.Pos2, ShouldEqual, "val2")
			So(mf.Sub.F1, ShouldEqual, 120)
		})
	})
}
