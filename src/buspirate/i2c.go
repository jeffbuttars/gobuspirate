
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

    // I2C_EXIT = 0x00
    // I2C_VERSION = 0x01
    I2C_SEND_START = 0x02
    I2C_SEND_STOP = 0x03
    I2C_READ_BYTE = 0x04
    I2C_SEND_ACK = 0x06
    I2C_SEND_NACK = 0x07
    // I2C_SNIFFER = 0x0F
    I2C_BULK_SEND = 0x10
    // I2C_SET_PERIPH = 0x40
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
}

func NewI2C(bp *BP) *I2C {
    return &I2C{Bp: bp}
} //NewI2C()


func (i2c *I2C) setPeriph(bit uint8, on bool) error {

    log.Printf("setPeriph %b\n", on)

    if on {
        return i2c.Bp.SetPinsIn(bit, string(0x01))
    }

    return i2c.Bp.SetPinsOut(bit, string(0x01))
} //setPeriph()

func (i2c *I2C) Power(on bool) error {
    log.Printf("Power %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_POWER, on)
} //Power()

func (i2c *I2C) Pullups(on bool) error {
    log.Printf("Pullups %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_PULLUPS, on)
} //Pullups()

func (i2c *I2C) AUX(on bool) error {
    log.Printf("AUX %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_AUX, on)
} //AUX()


func (i2c *I2C) CS(on bool) error {
    log.Printf("CS %b\n", on)
    return i2c.setPeriph(I2C_PERIPH_CS, on)
} //CS()

// func (i2c *I2C) Sniff() error {
// } //Sniff()

func (i2c *I2C) Start() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_START}, string(0x01))
    return err
} //Start()

func (i2c *I2C) Stop() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_STOP}, string(0x01))
    return err
} //Stop()

func (i2c *I2C) ACK() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_ACK}, string(0x01))
    return err
} //ACK()

func (i2c *I2C) NACK() error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SEND_NACK}, string(0x01))
    return err
} //NACK()

func (i2c *I2C) ReadByte() (uint8, error) {
    bytes, err := i2c.Bp.writeFind([]uint8{I2C_READ_BYTE}, "")
    return bytes[0], err
} //ReadByte()

func (i2c *I2C) setSpeed(speed uint8) error {
    _, err := i2c.Bp.writeFind([]uint8{I2C_SET_SPEED | speed}, string(0x01))
    if err != nil {
        return err
    }

    return nil
} //setSpeed()

func (i2c *I2C) SetSpeed5() error {
   return i2c.setSpeed(I2C_SPEED_5)
} //SetSpeed5()

func (i2c *I2C) SetSpeed50() error {
   return i2c.setSpeed(I2C_SPEED_50)
} //SetSpeed50()

func (i2c *I2C) SetSpeed100() error {
   return i2c.setSpeed(I2C_SPEED_100)
} //SetSpeed100()

func (i2c *I2C) SetSpeed400() error {
   return i2c.setSpeed(I2C_SPEED_400)
} //SetSpeed400()

func (i2c *I2C) SendBytes(bytes []uint8) error {

    if len(bytes) > 16 {
        return errors.New("Can't send more than 16 bytes at a time")
    }

    bp := i2c.Bp
    _, err := bp.Serial.Write([]uint8{I2C_BULK_SEND | uint8(len(bytes)-1)})
    if err != nil {
        return err
    }

    _, err = bp.writeFind(bytes, string(0x01))
    if err != nil {
        return err
    }

    return nil
} //SendBytes()
