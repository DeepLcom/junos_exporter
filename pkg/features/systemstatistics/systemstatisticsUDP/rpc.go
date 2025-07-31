package systemstatisticsUDP

import "encoding/xml"

type StatisticsUDP struct {
	XMLName    xml.Name `xml:"rpc-reply"`
	Text       string   `xml:",chardata"`
	Junos      string   `xml:"junos,attr"`
	Statistics struct {
		Text string `xml:",chardata"`
		Udp  struct {
			Text                                              string `xml:",chardata"`
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
		} `xml:"udp"`
	} `xml:"statistics"`
	Cli struct {
		Text   string `xml:",chardata"`
		Banner string `xml:"banner"`
	} `xml:"cli"`
}