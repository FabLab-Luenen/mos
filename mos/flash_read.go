//
// Copyright (c) 2014-2019 Cesanta Software Limited
// All rights reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// +build !noflash

package main

import (
	"io/ioutil"
	"os"
	"strconv"

	"context"

	"github.com/mongoose-os/mos/mos/dev"
	"github.com/mongoose-os/mos/mos/devutil"
	"github.com/mongoose-os/mos/mos/flags"
	"github.com/mongoose-os/mos/mos/flash/esp"
	espFlasher "github.com/mongoose-os/mos/mos/flash/esp/flasher"
	"github.com/cesanta/errors"
	flag "github.com/spf13/pflag"
)

func flashRead(ctx context.Context, devConn dev.DevConn) error {
	// if given devConn is not nil, we should disconnect it while flash reading is in progress
	if devConn != nil {
		devConn.Disconnect(ctx)
		defer devConn.Connect(ctx, true)
	}

	var err error
	var addr, length int64
	outFile := ""
	args := flag.Args()
	switch len(args) {
	case 2:
		// Nothing, will auto-detect the size and read entire flash.
		outFile = args[1]
	case 4:
		addr, err = strconv.ParseInt(args[1], 0, 64)
		if err != nil {
			return errors.Annotatef(err, "invalid address")
		}
		length, err = strconv.ParseInt(args[2], 0, 64)
		if err != nil {
			return errors.Annotatef(err, "invalid length")
		}
		outFile = args[3]
	default:
		return errors.Errorf("invalid arguments")
	}

	port, err := devutil.GetPort()
	if err != nil {
		return errors.Trace(err)
	}

	var data []byte
	platform := flags.Platform()
	switch platform {
	case "cc3200":
		err = errors.NotImplementedf("flash reading for %s", platform)
	case "esp32":
		espFlashOpts.ControlPort = port
		data, err = espFlasher.ReadFlash(esp.ChipESP32, uint32(addr), int(length), &espFlashOpts)
	case "esp8266":
		espFlashOpts.ControlPort = port
		data, err = espFlasher.ReadFlash(esp.ChipESP8266, uint32(addr), int(length), &espFlashOpts)
	case "stm32":
		err = errors.NotImplementedf("flash reading for %s", platform)
	default:
		err = errors.Errorf("unsupported platform '%s'", platform)
	}

	if err == nil {
		if outFile == "-" {
			_, err = os.Stdout.Write(data)
		} else {
			err = ioutil.WriteFile(outFile, data, 0644)
			if err == nil {
				reportf("Wrote %s", outFile)
			}
		}
	}

	return errors.Trace(err)
}
