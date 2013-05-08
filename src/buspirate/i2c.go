
package buspirate

import (
    "log"
    "errors"
    "fmt"
)

type I2C struct {
    Bp *BP
}

func NewI2C(dev string) *I2C {

    i2c := I2C{Bp: NewBP(dev)}

    return &i2c
} //NewI2C()

func (i2c *I2C) Init() error {

    bp := i2c.Bp

    err := bp.Init()
    if err != nil {
        return err
    }

    // err = bp.Reset()
    // if err != nil {
    //     return err
    // }

    err = bp.BinaryMode()
    if err != nil {
        return err
    }

    err = bp.WriteRead([]byte{0x02})
    if err != nil {
        return err
    }

    if fmt.Sprintf("%s", bp.buf[:4]) != "I2C1" {
        err = errors.New("Unable to enter I2C mode")
        log.Fatal(err)
        return err
    }

    log.Printf("Entered I2C mode.")
    return nil
} //Init()
