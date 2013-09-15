
package main

import (
    // "github.com/tarm/goserial"
    // "time"
    "log"
    "buspirate"
    "fmt"
    "pwm"
)

// const (
//     khz400 = 0x03
//     khz100 = 0x02
//     khz50 = 0x01
//     khz5 = 0x00
//     PIN_POWER = 0x08
//     PIN_PULLUPS = 0x04
//     PIN_AUX = 0x02
//     PIN_CS = 0x01
// )

func main() {

    bp := buspirate.NewBP("")
    err := bp.Init()
    if err != nil {
        log.Fatal(err)
    }

    //bp.Reset()
    bp.BinaryMode()


    m, err := bp.GetMode()
    fmt.Printf("MODE: %s\n", m)

    // bp.ShortTest()
    // bp.LongTest()

    i2c, err := bp.ModeI2C()
    if err != nil {
        log.Fatal(err)
   }

    i2c.Power(true)
    i2c.Pullups(true)
    // fmt.Printf("Starting scan...\n")
    // scan_res := i2c.Scan()
    // for _, v := range scan_res  {
    //     if (v & 0x01) == 0x01 {
    //         fmt.Printf("0x%2.2X(0x%2.2X R)\n",  v >> 1, v)
    //     } else {
    //         fmt.Printf("0x%2.2X(0x%2.2X W)\n", v >> 1, v)
    //     }
    // }

    // time.Sleep(300 * time.Millisecond)

    pwm.AllOn(i2c, 0, 4095)
    i2c.Power(false)
    i2c.Pullups(false)

    return
}
