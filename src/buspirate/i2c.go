
package buspirate

import (
    "log"
    "errors"
    // "fmt"
)

const (
    // http://dangerousprototypes.com/2009/10/14/bus-pirate-binary-i2c-mode/
    // 00000000 – Exit to bitbang mode, responds “BBIOx”
    // 00000001 – Mode version string (I2C1)
    // 00000010 – Send I2C start bit
    // 00000011 – Send I2C stop bit
    // 00000100 – I2C read byte
    // 00000110 – Send I2C ACK bit
    // 00000111 – Send I2C NACK bit
    // 00001111 – Start bus sniffer
    // 0001xxxx – Bulk transfer, send 1-16 bytes (0=1byte!)
    // 0100wxyz – Configure peripherals w=power, x=pullups, y=AUX, z=CS
    // 011000xx – Set I2C speed, 3=~400kHz, 2=~100kHz, 1=~50kHz, 0=~5kHz

    I2C_EXIT = 0x00
    I2C_VERSION = 0x01
    I2C_SEND_START = 0x02
    I2C_SEND_STOP = 0x03
    I2C_READ_BYTE = 0x04
    I2C_SEND_ACK = 0x06
    I2C_SEND_NACK = 0x07
    I2C_SNIFFER = 0x0F
    I2C_BULK_SEND = 0x10
    I2C_SET_PERIPH = 0x40
    I2C_PERIPH_POWER = 0x08
    I2C_PERIPH_PULLUPS = 0x04
    I2C_PERIPH_AUX = 0x02
    I2C_PERIPH_CS = 0x01
    I2C_SET_SPEED = 0x60
    I2C_SPEED_400 = 0x03
    I2C_SPEED_100 = 0x02
    I2C_SPEED_50 = 0x01
    I2C_SPEED_5 = 0x00
)

type I2C struct {
    Bp *BP
    periphs uint8
}

func NewI2C(dev string) *I2C {

    i2c := I2C{Bp: NewBP(dev), periphs: I2C_SET_PERIPH}

    return &i2c
} //NewI2C()

func (i2c *I2C) Mode() error {
    bp := i2c.Bp

    // log.Printf("The Buffer: %q\n", bp.buf)
    found, err := bp.WriteReadCHK([]byte{0x02}, "I2C1")
    if err != nil {
        return err
    }

    if found {
        log.Printf("Entered I2C mode.")
        return nil
    }

    // log.Printf("Unable to enter I2C mode, The Buffer: %q\n", bp.buf)
    err = errors.New("Unable to enter I2C mode")
    log.Fatal(err)

    return err
} //Mode()

func (i2c *I2C) Init() error {

    bp := i2c.Bp

    err := bp.Init()
    if err != nil {
        return err
    }

    err = bp.BinaryMode()
    if err != nil {
        return err
    }

    m_err := i2c.Mode()
    if m_err != nil {
        //Try again, but see if we can break out the current mode first.
        m_err = bp.Break()
        if m_err != nil {
            return m_err
        }

        m_err = bp.BinaryMode()
        if m_err != nil {
            return m_err
        }

        m_err = i2c.Mode()
    }

    if m_err != nil {
        log.Printf("Unable to enter I2C mode, The Buffer: %q\n", bp.buf)
        err = errors.New("Unable to enter I2C mode")
        log.Fatal(err)

        return m_err
    }


    return nil
} //Init()

func (i2c *I2C) periph(mask uint8, on bool) error {

    bp := i2c.Bp

    if on {
        i2c.periphs |= mask
    } else {
        i2c.periphs &= (0xFF ^ mask)
    }

    log.Printf("Periph: mask:%x, periphs: %x\n", mask, i2c.periphs)
    found, err := bp.WriteReadCHK([]byte{I2C_SET_PERIPH | i2c.periphs}, string(0x01))
    if err != nil {
        return err
    }

    if found {
        // log.Printf("Set peripheral mask %x\n", i2c.periphs)
        return nil
    }

    log.Printf("Unable set Peripheral mask: %x\n", i2c.periphs)
    err = errors.New("Unable set Peripheral mask")
    log.Fatal(err)

    return err
} //periph()

func (i2c *I2C) Power(on bool) error {
    log.Printf("Power\n")
    return i2c.periph(I2C_PERIPH_POWER, on)

    // } else {
    //     log.Printf("Power Off\n")
    //     // i2c.periphs &^= I2C_PERIPH_POWER
    //     return i2c.periph(0xFF ^ I2C_PERIPH_POWER)
    // }

    // return i2c.periph()
} //Power()
