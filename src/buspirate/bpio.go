
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
)

type BP struct {
	Device string
    Serial io.ReadWriteCloser
    SerialConf *serial.Config
    ReadTimeout time.Duration
    buf []uint8
}

func NewBP(dev string) *BP {

    if dev == "" {
       dev = "/dev/buspirate" 
    }

    bp := BP{Device: dev, ReadTimeout: 100 * time.Millisecond}

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

        log.Printf("WriteReadCHK comparing %q:%q\n", bp.buf[:len(chk)], chk)
        if fmt.Sprintf("%s", bp.buf[:len(chk)]) == chk {
            return true, nil
        }

        return false, nil
}

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

func (bp *BP) Reset() (bool, error) {
    log.Printf("Reset")

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
            return false, err
        }

        if found {
            log.Printf("Reset, 10 enters, Reset Good!")
            return true, nil
        } else {
            log.Printf("Reset, looking for %q, got %s:%q",
                        hiz, bp.buf, bp.buf)
        }
    }

    // Are we in binary mode?
    if fmt.Sprintf("%s", bp.buf[:5]) == "BBIO1" {
        log.Printf("Reset, 10 enters, Reset Good!")
        return true, nil
    }

    found, err := bp.WriteReadCHK([]uint8{'#', 0x0D}, hiz)

    if err != nil {
        log.Fatal(err)
        return false, err
    }
    if found {
        log.Printf("Reset, 10 enters, Reset Good!")
        return true, nil
    } else {
        log.Printf(
            "Reset #, looking for HiZ>, got %s:%q\n Don't expect this to work!\n",
            bp.buf, bp.buf)
    }

    // if fmt.Sprintf("%s", bp.buf[:5]) == "BBIO1" {
    //     log.Printf("Reset, 10 enters, in binmode")
    //     return true, nil
    // }

    return false, nil
} //Reset()

func (bp *BP) BinaryMode() error {

    rst, err := bp.Reset()
    if !rst {
        // Try to break and try again, this doesn't work yet :(
        bp.Break()
        rst, err = bp.Reset()
        if !rst {
            log.Printf("BinaryMode, unable to to reset BusPirate, not attempting Binary Mode\n")
            return nil
        }
    }

    // Enable Binary mode
    bm := []uint8{0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0 }
    bmi := "BBIO1"
    for k, _ := range bm {
        found, err := bp.WriteReadCHK(bm[k:k+1], bmi)

        if err != nil {
            log.Fatal(err)
            return err
        }

        if found {
            log.Printf("Entered Binary mode")
            return nil
        }
    }

    err = errors.New("Unable to enter binary mode")
    log.Fatal(err)
    return err
} //BinaryMode()

func (bp *BP) Break() error {

    log.Printf("Break()\n")

    bp.buf[0] = 0
    found, err := bp.WriteReadCHK([]byte{0x00},  "BBIO1")
    if err != nil {
        return err
    }

    if found {
        return nil
    }

    return err
} //Break()
