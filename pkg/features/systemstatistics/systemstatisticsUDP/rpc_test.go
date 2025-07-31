package systemstatisticsUDP

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

// UDPStats is a named struct for UDP statistics.
type UDPStats struct {
	Text                                              string  `xml:",chardata"`
	DatagramsReceived                                 float64 `xml:"datagrams-received"`
	DatagramsWithIncompleteHeader                     float64 `xml:"datagrams-with-incomplete-header"`
	DatagramsWithBadDatalengthField                   float64 `xml:"datagrams-with-bad-datalength-field"`
	DatagramsWithBadChecksum                          float64 `xml:"datagrams-with-bad-checksum"`
	DatagramsDroppedDueToNoSocket                     float64 `xml:"datagrams-dropped-due-to-no-socket"`
	BroadcastOrMulticastDatagramsDroppedDueToNoSocket float64 `xml:"broadcast-or-multicast-datagrams-dropped-due-to-no-socket"`
	DatagramsDroppedDueToFullSocketBuffers            float64 `xml:"datagrams-dropped-due-to-full-socket-buffers"`
	DatagramsNotForHashedPcb                          float64 `xml:"datagrams-not-for-hashed-pcb"`
	DatagramsDelivered                                float64 `xml:"datagrams-delivered"`
	DatagramsOutput                                   float64 `xml:"datagrams-output"`
}

type StatsUDP struct {
	XMLName xml.Name `xml:"statistics"`
	Text    string   `xml:",chardata"`
	Udp     UDPStats `xml:"udp"`
}

func TestUnmarshalUDPStatistics_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		xmlData  string
		expected StatsUDP
	}{
		{
			name: "Non-zero values",
			xmlData: `
<statistics>
    <udp>
        <datagrams-received>10</datagrams-received>
        <datagrams-with-incomplete-header>2</datagrams-with-incomplete-header>
        <datagrams-with-bad-datalength-field>1</datagrams-with-bad-datalength-field>
        <datagrams-with-bad-checksum>3</datagrams-with-bad-checksum>
        <datagrams-dropped-due-to-no-socket>4</datagrams-dropped-due-to-no-socket>
        <broadcast-or-multicast-datagrams-dropped-due-to-no-socket>5</broadcast-or-multicast-datagrams-dropped-due-to-no-socket>
        <datagrams-dropped-due-to-full-socket-buffers>6</datagrams-dropped-due-to-full-socket-buffers>
        <datagrams-not-for-hashed-pcb>7</datagrams-not-for-hashed-pcb>
        <datagrams-delivered>8</datagrams-delivered>
        <datagrams-output>9</datagrams-output>
    </udp>
</statistics>
`,
			expected: StatsUDP{
				Udp: UDPStats{
					DatagramsReceived:                                 10,
					DatagramsWithIncompleteHeader:                     2,
					DatagramsWithBadDatalengthField:                   1,
					DatagramsWithBadChecksum:                          3,
					DatagramsDroppedDueToNoSocket:                     4,
					BroadcastOrMulticastDatagramsDroppedDueToNoSocket: 5,
					DatagramsDroppedDueToFullSocketBuffers:            6,
					DatagramsNotForHashedPcb:                          7,
					DatagramsDelivered:                                8,
					DatagramsOutput:                                   9,
				},
			},
		},
		{
			name: "All zero values",
			xmlData: `
<statistics>
    <udp>
        <datagrams-received>0</datagrams-received>
        <datagrams-with-incomplete-header>0</datagrams-with-incomplete-header>
        <datagrams-with-bad-datalength-field>0</datagrams-with-bad-datalength-field>
        <datagrams-with-bad-checksum>0</datagrams-with-bad-checksum>
        <datagrams-dropped-due-to-no-socket>0</datagrams-dropped-due-to-no-socket>
        <broadcast-or-multicast-datagrams-dropped-due-to-no-socket>0</broadcast-or-multicast-datagrams-dropped-due-to-no-socket>
        <datagrams-dropped-due-to-full-socket-buffers>0</datagrams-dropped-due-to-full-socket-buffers>
        <datagrams-not-for-hashed-pcb>0</datagrams-not-for-hashed-pcb>
        <datagrams-delivered>0</datagrams-delivered>
        <datagrams-output>0</datagrams-output>
    </udp>
</statistics>
`,
			expected: StatsUDP{
				Udp: UDPStats{
					DatagramsReceived:                                 0,
					DatagramsWithIncompleteHeader:                     0,
					DatagramsWithBadDatalengthField:                   0,
					DatagramsWithBadChecksum:                          0,
					DatagramsDroppedDueToNoSocket:                     0,
					BroadcastOrMulticastDatagramsDroppedDueToNoSocket: 0,
					DatagramsDroppedDueToFullSocketBuffers:            0,
					DatagramsNotForHashedPcb:                          0,
					DatagramsDelivered:                                0,
					DatagramsOutput:                                   0,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual StatsUDP
			err := xml.Unmarshal([]byte(tc.xmlData), &actual)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			assert.Equal(t, tc.expected.Udp.DatagramsReceived, actual.Udp.DatagramsReceived)
			assert.Equal(t, tc.expected.Udp.DatagramsWithIncompleteHeader, actual.Udp.DatagramsWithIncompleteHeader)
			assert.Equal(t, tc.expected.Udp.DatagramsWithBadDatalengthField, actual.Udp.DatagramsWithBadDatalengthField)
			assert.Equal(t, tc.expected.Udp.DatagramsWithBadChecksum, actual.Udp.DatagramsWithBadChecksum)
			assert.Equal(t, tc.expected.Udp.DatagramsDroppedDueToNoSocket, actual.Udp.DatagramsDroppedDueToNoSocket)
			assert.Equal(t, tc.expected.Udp.BroadcastOrMulticastDatagramsDroppedDueToNoSocket, actual.Udp.BroadcastOrMulticastDatagramsDroppedDueToNoSocket)
			assert.Equal(t, tc.expected.Udp.DatagramsDroppedDueToFullSocketBuffers, actual.Udp.DatagramsDroppedDueToFullSocketBuffers)
			assert.Equal(t, tc.expected.Udp.DatagramsNotForHashedPcb, actual.Udp.DatagramsNotForHashedPcb)
			assert.Equal(t, tc.expected.Udp.DatagramsDelivered, actual.Udp.DatagramsDelivered)
			assert.Equal(t, tc.expected.Udp.DatagramsOutput, actual.Udp.DatagramsOutput)
		})
	}
}