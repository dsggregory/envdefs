# Default Struct From Environment
This package allows one to define a Golang structure with tags so the struct pointer can set defaults based on tagged environment variables.

## Usage

```go
package main

import (
	"log"
	"os"
)
import "time"
import "dsggregory/envdefs"

type MyStruct struct {
	OutOfDateDur  time.Duration `env:"OUT_OF_DATE_DUR"`
	SourceDir     string        `env:"SOURCE_DIR,required"`
	AgentConfFile string        `env:"AGENT_CONFIG_FILE,required"`

	DryRun bool `env:"DRY_RUN"`
}

func main() {
	ms := MyStruct{
		OutOfDateDur: time.hour * 24, // can be overridden by env var
	}
	if err := envdefs.ReadDefaults(&ms); err != nil {
		// missing required env vars returns an error
		log.Fatal(err)
	}
	os.Chdir(ms.SourceDir)
	...
}
```