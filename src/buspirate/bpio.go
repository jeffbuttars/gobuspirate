
package buspirate

import (
    "github.com/tarm/goserial"
    "io"
    "log"
    "time"
    "fmt"
    "errors"
)

const (
    BAUD = 115200
    BUF_SIZE = 1024
    HW_RESET = 0x0F
    HW_RESET_REPLY = 0x01
    BINARY_RESET = 0x00
    STATE_INITIAL = 0x01
    STATE_BINARY = 0x02
    STATE_RAW = 0x04

    MODE_SPI = 0x01
    MODE_I2C = 0x02
    MODE_I2C_REPLY = "I2C1"
    MODE_UART = 0x03
    MODE_1WIRE = 0x04
    MODE_RAW = 0x05
    GET_MODE = 0x01

    HW_TEST_SHORT = 0x10
    HW_TEST_LONG = 0x11

    SET_PWM = 0x12
    CLEAR_PWM = 0x13

    VOLT_MEASURE = 0x14
    SET_PINS_IN_OUT = 0x40
    SET_PINS_HIGH_LOW = 0x80

    DEFAULT_TIMEOUT = 100
)

type BP struct {
	Device string
    Serial io.ReadWriteCloser
    SerialConf *serial.Config
    ReadTimeout time.Duration
    buf []uint8
    pins_high_low uint8
    pins_in_out uint8
    state uint8
}

func NewBP(dev string) *BP {

    if dev == "" {
       dev = "/dev/buspirate" 
    }

    bp := BP{Device: dev, ReadTimeout: DEFAULT_TIMEOUT * time.Millisecond,
            pins_high_low: SET_PINS_HIGH_LOW,
            pins_in_out: SET_PINS_IN_OUT,
        }

    return &bp
} //NewBP()

func (bp *BP) Init() error {

    bp.SerialConf = &serial.Config{Name: bp.Device, Baud: BAUD}
    s, err := serial.OpenPort(bp.SerialConf)

    if err != nil {
        return err
    }

    bp.Serial = s

    bp.buf = make([]uint8, BUF_SIZE)

    return nil
} //Init()

func (bp *BP) ReadNB() (int, error) {
        // return 0, errors.New("Unable to enter binary mode, Fucker!!!")

        fr := make(chan int)  // File Result, bytes read
        er := make(chan error)  // Errer number

        var fres int
        var ferr error

        go func() {
            n, err := bp.Serial.Read(bp.buf)
            if err != nil {
                er <- err
                return
            }
            // time.Sleep(1000 * time.Millisecond)
            fmt.Printf("Read %d bytes\n", n)
            fr <- n
            fmt.Printf("ReadNB done\n", n)
        }()

        // Try to read a result right away,
        // if we don't have one, wait for the timeout
        // period and try again. If we have no data and no error,
        // we give up.
        fmt.Printf("ReadNB Select 1\n")
        select {
            case fres = <-fr:
                return fres, nil
            case ferr = <-er:
                return 0, ferr
            default:
                // fmt.Printf("ReadNB Select 1 Sleeping %d\n", bp.ReadTimeout)
                time.Sleep(bp.ReadTimeout)
        }

        fmt.Printf("ReadNB Select 2\n")
        select {
            case fres = <-fr:
                return fres, nil
            case ferr = <-er:
                return 0, ferr
            default:
                return 0, nil
        }

} //ReadNB()


// Every read is checked for a match of 'chk', returns true/false for check
// result
func (bp *BP) WriteReadCHK(data []uint8, chk string) (bool, error) {

        err := bp.WriteRead(data)
        if err != nil {
            log.Fatal(err)
            return false, err
        }

        // If chk is empty, always return a find.
        if len(chk) == 0 {
            return true, nil
        }

        log.Printf("WriteReadCHK comparing %q:%q\n", bp.buf[:len(chk)], chk)
        if fmt.Sprintf("%s", bp.buf[:len(chk)]) == chk {
            return true, nil
        }

        return false, nil
}

func (bp *BP) writeFind(data []uint8, chk string) error {
    // A wrapper to simplify use of WriteReadCHK

    found, err := bp.WriteReadCHK(data, chk)
    if err != nil {
        return err
    }

    if found {
        return nil
    }

    return errors.New(fmt.Sprintf("Unable to find chk string: %q", chk))
} //writeFind()

func (bp *BP) WriteRead(data []uint8) error {

    fmt.Printf("WriteRead, writing: %q\n", data)

    n, err := bp.Serial.Write(data)
    if err != nil {
        log.Fatal(err)
        return err
    }

    n, err = bp.ReadNB()
    if err != nil {
        log.Fatal(err)
        return err
    }

    log.Printf("WriteRead read: %d:%s\n", n, bp.buf[:n])

    return nil
} //WriteRead()

func (bp *BP) HWReset() error {
    // We use a long time to give the board plenty of reset time.
    bp.ReadTimeout = 500 * time.Millisecond
    err := bp.writeFind([]uint8{HW_RESET}, string(HW_RESET_REPLY))
    bp.ReadTimeout = DEFAULT_TIMEOUT * time.Millisecond

    if err == nil {
        bp.state = STATE_INITIAL
    }

    if string(bp.buf[0]) == "\a" {
        // We might already be at console mode, let's s
        log.Printf("HWReset: attempting normal reset.")
        bp.state = STATE_INITIAL
        return bp.Reset()
    }

    return err
} //HWReset()

func (bp *BP) Reset() error {
    log.Printf("Reset")

    // Are we already Reset?
    if bp.state == STATE_INITIAL {
        log.Printf("Already in intial state, look for HiZ> anyway")
    }

    if bp.state == STATE_BINARY {
        log.Printf("In binary mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    if bp.state == STATE_RAW {
        log.Printf("In raw mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    if bp.state == 0 {
        log.Printf("In UNKNOWN mode state, going to reset HW")
        err := bp.HWReset()
        if err != nil {
            log.Printf("Reset: Unable to Hardware reset.")
            return err
        }
    }

    // Reset the BP
    // Send 10 <enter> and one '#'
    rst_bits := []uint8{
        0x0D, 0x0D, 0x0D, 0x0D, 0x0D,
        0x0D, 0x0D, 0x0D, 0x0D, 0x0D }

    hiz := "\r\nHiZ>"
    for k, _ := range rst_bits {

        found, err := bp.WriteReadCHK(rst_bits[k:k+1], hiz)

        if err != nil {
            log.Fatal(err)
            return err
        }

        if found {
            log.Printf("Reset, 10 enters, Reset Good!")
            bp.state = STATE_INITIAL
            return nil
        } else {
            log.Printf("Reset, looking for %q, got %s:%q",
                        hiz, bp.buf, bp.buf)
        }
    }

    // Are we in binary mode?
    if fmt.Sprintf("%s", bp.buf[:5]) == "BBIO1" {
        log.Printf("Reset, 10 enters, Reset Good!")
        bp.state = STATE_INITIAL
        return nil
    }

    err := bp.writeFind([]uint8{'#', 0x0D}, hiz)

    if err != nil {
        log.Fatal(err)
        log.Printf(
            "Reset #, looking for HiZ>, got %s:%q\n Don't expect this to work!\n",
            bp.buf, bp.buf)
        return err
    }

    log.Printf("Reset, 10 enters, Reset Good!")
    bp.state = STATE_INITIAL
    return nil
} //Reset()

func (bp *BP) BinaryMode() error {

    if bp.state == STATE_BINARY {
        log.Printf("BinaryMode: Already in binary mode, not going to reset.")
    } else {
        log.Printf("BinaryMode: Reseting...")
        err := bp.Reset()
        if err != nil {
            log.Printf(
                "BinaryMode, unable to to reset BusPirate, not attempting Binary Mode\n")
            return err
        }
    }

    // Enable Binary mode
    bm := []uint8{0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0 }
    bmi := "BBIO1"
    for k, _ := range bm {
        err := bp.writeFind(bm[k:k+1], bmi)
        if err == nil {
            log.Printf("Entered Binary mode")
            bp.state = STATE_BINARY
            return nil
        }
    }

    err := errors.New("Unable to enter binary mode")
    log.Fatal(err)
    return err
} //BinaryMode()

// func (bp *BP) Break() error {

//     log.Printf("Break()\n")

//     bp.buf[0] = 0
//     found, err := bp.WriteReadCHK([]byte{0x00},  "BBIO1")
//     if err != nil {
//         return err
//     }

//     if found {
//         return nil
//     }

//     return err
// } //Break()

func (bp *BP) writePinsHL(chk string) error {
    log.Printf("writePinsHL: pins_high_low: %x\n", bp.pins_high_low)
    err := bp.writeFind([]byte{SET_PINS_HIGH_LOW | bp.pins_high_low}, chk)
    if err != nil {
        log.Printf("Unable set Peripheral mask: %x\n", bp.pins_high_low)
        return err
    }

    return nil
} //writePinsHL(chk string)()

func (bp *BP) writePinsIO(chk string) error {
    log.Printf("writePinsIO: pins_in_out: %x\n", bp.pins_in_out)
    err := bp.writeFind([]byte{SET_PINS_IN_OUT | bp.pins_in_out}, chk)
    if err != nil {
        log.Printf("Unable set Peripheral mask: %x\n", bp.pins_in_out)
        return err
    }

    return nil
} //writePinsIO(chk string)()

func (bp *BP) SetPinsHigh(mask uint8, chk string) error {

    bp.pins_high_low |= mask
    err := bp.writePinsHL(chk)
    if err != nil {
        bp.pins_high_low &= (0xFF ^ mask)
        return err
    }

    return nil
} //SetPinsHigh()

func (bp *BP) SetPinsLow(mask uint8, chk string) error {

    bp.pins_high_low &= (0xFF ^ mask)
    err := bp.writePinsHL(chk)
    if err != nil {
        bp.pins_high_low |= mask
        return err
    }

    return nil
} //SetPinsLow()

func (bp *BP) SetPinsIn(mask uint8, chk string) error {

    bp.pins_in_out |= mask
    err := bp.writePinsIO(chk)
    if err != nil {
        bp.pins_in_out &= (0xFF ^ mask)
        return err
    }

    return nil
} //SetPinsIn()

func (bp *BP) SetPinsOut(mask uint8, chk string) error {

    bp.pins_in_out &= (0xFF ^ mask)
    err := bp.writePinsIO(chk)
    if err != nil {
        bp.pins_in_out |= mask
        return err
    }

    return nil
} //SetPinsOut()

func (bp *BP) ShortTest() error {

    log.Printf("ShortTest")
    // HW_TEST_SHORT
    err := bp.writeFind([]byte{HW_TEST_SHORT}, "")
    fmt.Printf("ShortTest output: %s\n", bp.buf)

    return err
} //ShortTest()

func (bp *BP) LongTest() error {

    log.Printf("LongTest")
    // HW_TEST_SHORT
    err := bp.writeFind([]byte{HW_TEST_LONG}, "")
    fmt.Printf("LongTest output: %s\n", bp.buf)

    return err
} //LongTest()

func (bp *BP) GetMode() (string, error) {

    err := bp.writeFind([]byte{GET_MODE}, "")
    res := make([]uint8, len(bp.buf))
    copy(res, bp.buf)

    return string(res), err
} //GetMode()

func (bp *BP) ModeI2C() (*I2C, error) {

    i2c := NewI2C(bp)
    // i2c.Init()
    // Make sure we're in Binary Mode
    bp.BinaryMode()

    // log.Printf("The Buffer: %q\n", bp.buf)
    err := bp.writeFind([]byte{MODE_I2C}, MODE_I2C_REPLY)
    if err != nil {
        return i2c, err
    }

    log.Printf("Entered I2C mode.")
    return i2c, nil
} //ModeI2C()
