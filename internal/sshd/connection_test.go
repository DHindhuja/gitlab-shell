package sshd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

type rejectCall struct {
	reason  ssh.RejectionReason
	message string
}

type fakeNewChannel struct {
	channelType string
	extraData   []byte
	acceptErr   error

	acceptCh chan struct{}
	rejectCh chan rejectCall
}

func (f *fakeNewChannel) Accept() (ssh.Channel, <-chan *ssh.Request, error) {
	if f.acceptCh != nil {
		f.acceptCh <- struct{}{}
	}

	return nil, nil, f.acceptErr
}

func (f *fakeNewChannel) Reject(reason ssh.RejectionReason, message string) error {
	if f.rejectCh != nil {
		f.rejectCh <- rejectCall{reason: reason, message: message}
	}

	return nil
}

func (f *fakeNewChannel) ChannelType() string {
	return f.channelType
}

func (f *fakeNewChannel) ExtraData() []byte {
	return f.extraData
}

func setup(sessionsNum int64, newChannel *fakeNewChannel) (*connection, chan ssh.NewChannel) {
	conn := newConnection(sessionsNum, "127.0.0.1:50000")

	chans := make(chan ssh.NewChannel, 1)
	chans <- newChannel

	return conn, chans
}

func TestPanicDuringSessionIsRecovered(t *testing.T) {
	newChannel := &fakeNewChannel{channelType: "session"}
	conn, chans := setup(1, newChannel)

	numSessions := 0
	require.NotPanics(t, func() {
		conn.handle(context.Background(), chans, func(context.Context, ssh.Channel, <-chan *ssh.Request) {
			numSessions += 1
			close(chans)
			panic("This is a panic")
		})
	})

	require.Equal(t, numSessions, 1)
}

func TestUnknownChannelType(t *testing.T) {
	rejectCh := make(chan rejectCall)
	defer close(rejectCh)

	newChannel := &fakeNewChannel{channelType: "unknown session", rejectCh: rejectCh}
	conn, chans := setup(1, newChannel)

	go func() {
		conn.handle(context.Background(), chans, nil)
	}()

	rejectionData := <-rejectCh

	expectedRejection := rejectCall{reason: ssh.UnknownChannelType, message: "unknown channel type"}
	require.Equal(t, expectedRejection, rejectionData)
}

func TestTooManySessions(t *testing.T) {
	rejectCh := make(chan rejectCall)
	defer close(rejectCh)

	newChannel := &fakeNewChannel{channelType: "session", rejectCh: rejectCh}
	conn, chans := setup(1, newChannel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		conn.handle(context.Background(), chans, func(context.Context, ssh.Channel, <-chan *ssh.Request) {
			<-ctx.Done() // Keep the accepted channel open until the end of the test
		})
	}()

	chans <- newChannel
	require.Equal(t, <-rejectCh, rejectCall{reason: ssh.ResourceShortage, message: "too many concurrent sessions"})
}

func TestAcceptSessionSucceeds(t *testing.T) {
	newChannel := &fakeNewChannel{channelType: "session"}
	conn, chans := setup(1, newChannel)

	channelHandled := false
	conn.handle(context.Background(), chans, func(context.Context, ssh.Channel, <-chan *ssh.Request) {
		channelHandled = true
		close(chans)
	})

	require.True(t, channelHandled)
}

func TestAcceptSessionFails(t *testing.T) {
	acceptCh := make(chan struct{})
	defer close(acceptCh)

	acceptErr := errors.New("some failure")
	newChannel := &fakeNewChannel{channelType: "session", acceptCh: acceptCh, acceptErr: acceptErr}
	conn, chans := setup(1, newChannel)

	channelHandled := false
	go func() {
		conn.handle(context.Background(), chans, func(context.Context, ssh.Channel, <-chan *ssh.Request) {
			channelHandled = true
		})
	}()

	require.Equal(t, <-acceptCh, struct{}{})

	// Waits until the number of sessions is back to 0, since we can only have 1
	conn.concurrentSessions.Acquire(context.Background(), 1)
	defer conn.concurrentSessions.Release(1)

	require.False(t, channelHandled)
}
