package defaults

import (
	"os"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type testData struct {
	F1   string `env:"F1"`
	Home string `env:"HOME"`
	UID  int    `env:"UID"`
	User string `env:"USER"`
}

func (td *testData) export() {
	if td.F1 != "" {
		_ = os.Setenv("F1", td.F1)
	}
	if td.Home != "" {
		_ = os.Setenv("HOME", td.Home)
	}
	_ = os.Setenv("UID", strconv.Itoa(td.UID))
	if td.User != "" {
		_ = os.Setenv("USER", td.User)
	}
}

func TestOne(t *testing.T) {
	Convey("Empty defaults", t, func() {
		td := testData{
			F1: "field1",
		}
		err := ReadDefaults(&td)
		So(err, ShouldBeNil)
		So(td.UID, ShouldEqual, 0)
	})
	Convey("Defaults", t, func() {
		tdEnv := testData{
			UID: 50,
		}
		tdEnv.export()
		td := testData{
			F1: "field1",
		}
		err := ReadDefaults(&td)
		So(err, ShouldBeNil)
		So(td.UID, ShouldEqual, 50)
	})
	Convey("Required", t, func() {
		type t struct {
			Need string `env:"NEED_IT,required"`
		}
		envName := "NEED_IT"
		expVal := "set"

		td := t{}
		err := ReadDefaults(&td)
		So(err, ShouldNotBeNil)

		_ = os.Setenv(envName, expVal)
		err = ReadDefaults(&td)
		So(err, ShouldBeNil)
		So(td.Need, ShouldEqual, expVal)

		td.Need = expVal
		_ = os.Unsetenv(envName)
		err = ReadDefaults(&td)
		So(err, ShouldBeNil)
		So(td.Need, ShouldEqual, expVal)
	})
}
