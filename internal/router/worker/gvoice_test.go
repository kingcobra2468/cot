package worker

import (
	"testing"
	"time"

	"github.com/kingcobra2468/cot/internal/router/worker/gvoice"
	"github.com/kingcobra2468/cot/internal/router/worker/gvoice/mocks"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func unixTime() uint64 {
	return uint64(time.Now().Unix()) * 1000
}

func textsEqual(t *testing.T, expected, actual *[]Text) {
	for i := range *expected {
		assert.Equal(t, (*expected)[i].Message, (*actual)[i].Message, "messages not equal")
		assert.Equal(t, (*expected)[i].Timestamp, (*actual)[i].Timestamp, "timestamps not equal")
	}
}

var hour uint64 = uint64(time.Hour.Microseconds())
var futureTime uint64 = unixTime() + hour
var pastTime uint64 = unixTime() - hour

func TestNewTexts(t *testing.T) {
	var tests = []struct {
		name  string
		input []*gvoice.MessageNode
		want  *[]Text
	}{
		{"0 Messages", []*gvoice.MessageNode{}, &[]Text{}},
		{"1 Message", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test"), Source: lo.ToPtr(true)}}, &[]Text{{Message: "test", Timestamp: futureTime}}},
		{"2 Messages", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test"), Source: lo.ToPtr(true)}, {Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test1"), Source: lo.ToPtr(true)}}, &[]Text{{Message: "test", Timestamp: futureTime}, {Message: "test1", Timestamp: futureTime}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGVoiceClient := mocks.NewGVoiceClient(t)
			mockGVoiceClient.EXPECT().GetContactHistory(mock.Anything, mock.Anything, mock.Anything).Return(&gvoice.FetchContactHistoryResponse{Success: lo.ToPtr(true), Messages: tt.input}, nil)
			w := NewGVoiceWorker(Link{"", ""}, false, mockGVoiceClient)

			texts, _ := w.newTexts(10)
			assert.Equal(t, len(*texts), len(*tt.want))
			textsEqual(t, texts, tt.want)
			mockGVoiceClient.AssertExpectations(t)
		})
	}
}

func TestUnprocessedTexts(t *testing.T) {
	var tests = []struct {
		name  string
		input []*gvoice.MessageNode
		want  *[]Text
	}{
		{"0 Messages", []*gvoice.MessageNode{}, &[]Text{}},
		{"1 Message", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test"), Source: lo.ToPtr(true)}}, &[]Text{{Message: "test", Timestamp: futureTime}}},
		{"2 Messages", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test"), Source: lo.ToPtr(true)}, {Timestamp: lo.ToPtr(pastTime), MessageContents: lo.ToPtr("test1"), Source: lo.ToPtr(true)}}, &[]Text{{Message: "test", Timestamp: futureTime}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGVoiceClient := mocks.NewGVoiceClient(t)
			mockGVoiceClient.EXPECT().GetContactHistory(mock.Anything, mock.Anything, mock.Anything).Return(&gvoice.FetchContactHistoryResponse{Success: lo.ToPtr(true), Messages: tt.input}, nil)
			w := NewGVoiceWorker(Link{"", ""}, false, mockGVoiceClient)

			texts, _ := w.unprocessedTexts()
			assert.Equal(t, len(*texts), len(*tt.want))
			textsEqual(t, texts, tt.want)
			mockGVoiceClient.AssertExpectations(t)
		})
	}
}

func TestFetch(t *testing.T) {
	var tests = []struct {
		name  string
		input []*gvoice.MessageNode
		want  *[]service.UserInput
	}{
		{"0 Messages", []*gvoice.MessageNode{}, &[]service.UserInput{}},
		{"1 Message", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test"), Source: lo.ToPtr(true)}}, &[]service.UserInput{{Name: "test", Args: []string{}, Raw: "test"}}},
		{"2 Messages", []*gvoice.MessageNode{{Timestamp: lo.ToPtr(futureTime), MessageContents: lo.ToPtr("test a"), Source: lo.ToPtr(true)}, {Timestamp: lo.ToPtr(pastTime), MessageContents: lo.ToPtr("test1"), Source: lo.ToPtr(true)}}, &[]service.UserInput{{Name: "test", Args: []string{"a"}, Raw: "test a"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGVoiceClient := mocks.NewGVoiceClient(t)
			mockGVoiceClient.EXPECT().GetContactHistory(mock.Anything, mock.Anything, mock.Anything).Return(&gvoice.FetchContactHistoryResponse{Success: lo.ToPtr(true), Messages: tt.input}, nil)
			w := NewGVoiceWorker(Link{"", ""}, false, mockGVoiceClient)

			ui := w.Fetch()
			assert.Equal(t, len(*ui), len(*tt.want))
			assert.ElementsMatch(t, *ui, *tt.want)
			mockGVoiceClient.AssertExpectations(t)
		})
	}
}
