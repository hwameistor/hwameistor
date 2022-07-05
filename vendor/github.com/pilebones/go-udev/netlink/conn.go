package netlink

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

type Mode int

// Mode determines event source: kernel events or udev-processed events.
// See libudev/libudev-monitor.c.
const (
	KernelEvent Mode = 1
	// Events that are processed by udev - much richer, with more attributes (such as vendor info, serial numbers and more).
	UdevEvent Mode = 2
)

// Generic connection
type NetlinkConn struct {
	Fd   int
	Addr syscall.SockaddrNetlink
}

type UEventConn struct {
	NetlinkConn

	// Options
	MatchedUEventLimit int // allow to stop monitor mode after X event(s) matched by the matcher
}

// Connect allow to connect to system socket AF_NETLINK with family NETLINK_KOBJECT_UEVENT to
// catch events about block/char device
// see:
// - http://elixir.free-electrons.com/linux/v3.12/source/include/uapi/linux/netlink.h#L23
// - http://elixir.free-electrons.com/linux/v3.12/source/include/uapi/linux/socket.h#L11
func (c *UEventConn) Connect(mode Mode) (err error) {

	if c.Fd, err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_KOBJECT_UEVENT); err != nil {
		return
	}

	c.Addr = syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: uint32(mode),
	}

	if err = syscall.Bind(c.Fd, &c.Addr); err != nil {
		syscall.Close(c.Fd)
	}

	return
}

// Close allow to close file descriptor and socket bound
func (c *UEventConn) Close() error {
	return syscall.Close(c.Fd)
}

func (c *UEventConn) msgPeek() (int, *[]byte, error) {
	var n int
	var err error
	buf := make([]byte, os.Getpagesize())
	for {
		// Just read how many bytes are available in the socket
		// Warning: syscall.MSG_PEEK is a blocking call
		if n, _, err = syscall.Recvfrom(c.Fd, buf, syscall.MSG_PEEK); err != nil {
			return n, &buf, err
		}

		// If all message could be store inside the buffer : break
		if n < len(buf) {
			break
		}

		// Increase size of buffer if not enough
		buf = make([]byte, len(buf)+os.Getpagesize())
	}
	return n, &buf, err
}

func (c *UEventConn) msgRead(buf *[]byte) error {
	if buf == nil {
		return errors.New("empty buffer")
	}

	n, _, err := syscall.Recvfrom(c.Fd, *buf, 0)
	if err != nil {
		return err
	}

	// Extract only real data from buffer and return that
	*buf = (*buf)[:n]

	return nil
}

// ReadMsg allow to read an entire uevent msg
func (c *UEventConn) ReadMsg() (msg []byte, err error) {
	// Just read how many bytes are available in the socket
	_, buf, err := c.msgPeek()
	if err != nil {
		return nil, err
	}

	// Now read complete data
	err = c.msgRead(buf)

	return *buf, err
}

// ReadMsg allow to read an entire uevent msg
func (c *UEventConn) ReadUEvent() (*UEvent, error) {
	msg, err := c.ReadMsg()
	if err != nil {
		return nil, err
	}

	return ParseUEvent(msg)
}

// Monitor run in background a worker to read netlink msg in loop and notify
// when msg receive inside a queue using channel.
// To be notified with only relevant message, use Matcher.
func (c *UEventConn) Monitor(queue chan UEvent, errs chan error, matcher Matcher) chan struct{} {
	quit := make(chan struct{}, 1)
	if matcher != nil {
		if err := matcher.Compile(); err != nil {
			errs <- fmt.Errorf("Wrong matcher, err: %w", err)
			quit <- struct{}{}
			close(queue)
			return quit
		}
	}

	go func() {
		bufToRead := make(chan *[]byte, 1)
		count := 0
	loop:
		for {
			select {
			case <-quit:
				break loop // stop iteration in case of stop signal received
			case buf := <-bufToRead: // Read one by one
				err := c.msgRead(buf)
				if err != nil {
					errs <- fmt.Errorf("Unable to read uevent, err: %w", err)
					break loop // stop iteration in case of error
				}

				uevent, err := ParseUEvent(*buf)
				if err != nil {
					errs <- fmt.Errorf("Unable to parse uevent, err: %w", err)
					continue loop // Drop uevent if not known
				}

				if matcher != nil {
					if !matcher.Evaluate(*uevent) {
						continue loop // Drop uevent if not match
					}
				}
				queue <- *uevent
				count++
				if c.MatchedUEventLimit > 0 && count >= c.MatchedUEventLimit {
					break loop // stop iteration when reach limit of uevent
				}
			default:
				_, buf, err := c.msgPeek()
				if err != nil {
					errs <- fmt.Errorf("Unable to check available uevent, err: %w", err)
					break loop // stop iteration in case of error
				}
				bufToRead <- buf
			}
		}
	}()
	return quit
}
