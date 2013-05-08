
package main

import (
    // "github.com/tarm/goserial"
    // "log"
    // "time"
    "buspirate"
)

const (
    khz400 = 0x03
    khz100 = 0x02
    khz50 = 0x01
    khz5 = 0x00
    PIN_POWER = 0x08
    PIN_PULLUPS = 0x04
    PIN_AUX = 0x02
    PIN_CS = 0x01
)

func main() {

    // bp := buspirate.NewBP("")
    // err := bp.Init()
    // if err != nil {
    //     log.Fatal(err)
    // }

    // bp.Reset()
    // bp.BinaryMode()

    i2c := buspirate.NewI2C("")
    i2c.Init()
    return
}
