package systemstatisticsUDP

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/czerwonk/junos_exporter/pkg/collector"
)

const prefix string = "junos_systemstatistics_udp_"

// Metrics to collect for the feature
var (
	datagramsReceivedDesc                                 *prometheus.Desc
	datagramsWithIncompleteHeaderDesc                     *prometheus.Desc
	datagramsWithBadDatalengthFieldDesc                   *prometheus.Desc
	datagramsWithBadChecksumDesc                          *prometheus.Desc
	datagramsDroppedDueToNoSocketDesc                     *prometheus.Desc
	broadcastOrMulticastDatagramsDroppedDueToNoSocketDesc *prometheus.Desc
	datagramsDroppedDueToFullSocketBuffersDesc            *prometheus.Desc
	datagramsNotForHashedPcbDesc                          *prometheus.Desc
	datagramsDeliveredDesc                                *prometheus.Desc
	datagramsOutputDesc                                   *prometheus.Desc
)

func init() {
	l := []string{"target", "protocol"}
	datagramsReceivedDesc = prometheus.NewDesc(prefix+"datagrams_received", "Number of UDP datagrams received", l, nil)
	datagramsWithIncompleteHeaderDesc = prometheus.NewDesc(prefix+"datagrams_with_incomplete_header", "Number of UDP datagrams with incomplete header", l, nil)
	datagramsWithBadDatalengthFieldDesc = prometheus.NewDesc(prefix+"datagrams_with_bad_datalength_field", "Number of UDP datagrams with bad datalength field", l, nil)
	datagramsWithBadChecksumDesc = prometheus.NewDesc(prefix+"datagrams_with_bad_checksum", "Number of UDP datagrams with bad checksum", l, nil)
	datagramsDroppedDueToNoSocketDesc = prometheus.NewDesc(prefix+"datagrams_dropped_due_to_no_socket", "Number of UDP datagrams dropped due to no socket", l, nil)
	broadcastOrMulticastDatagramsDroppedDueToNoSocketDesc = prometheus.NewDesc(prefix+"broadcast_or_multicast_datagrams_dropped_due_to_no_socket", "Number of UDP broadcast or multicast datagrams dropped due to no socket", l, nil)
	datagramsDroppedDueToFullSocketBuffersDesc = prometheus.NewDesc(prefix+"datagrams_dropped_due_to_full_socket_buffers", "Number of UDP datagrams dropped due to full socket buffers", l, nil)
	datagramsNotForHashedPcbDesc = prometheus.NewDesc(prefix+"datagrams_not_for_hashed_pcb", "Number of UDP datagrams not for hashed pcb", l, nil)
	datagramsDeliveredDesc = prometheus.NewDesc(prefix+"datagrams_delivered", "Number of UDP datagrams delivered", l, nil)
	datagramsOutputDesc = prometheus.NewDesc(prefix+"datagrams_output", "Number of UDP datagrams output", l, nil)
}

type systemstatisticsUDPCollector struct{}

func NewCollector() collector.RPCCollector {
	return &systemstatisticsUDPCollector{}
}

func (c *systemstatisticsUDPCollector) Name() string {
	return "systemstatisticsUDP"
}

func (c *systemstatisticsUDPCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- datagramsReceivedDesc
	ch <- datagramsWithIncompleteHeaderDesc
	ch <- datagramsWithBadDatalengthFieldDesc
	ch <- datagramsWithBadChecksumDesc
	ch <- datagramsDroppedDueToNoSocketDesc
	ch <- broadcastOrMulticastDatagramsDroppedDueToNoSocketDesc
	ch <- datagramsDroppedDueToFullSocketBuffersDesc
	ch <- datagramsNotForHashedPcbDesc
	ch <- datagramsDeliveredDesc
	ch <- datagramsOutputDesc
}

func (c *systemstatisticsUDPCollector) Collect(client collector.Client, ch chan<- prometheus.Metric, labelValues []string) error {
	var s StatisticsUDP
	err := client.RunCommandAndParse("show system statistics udp", &s)
	if err != nil {
		return err
	}
	c.collectSystemStatisticsUDP(ch, labelValues, s)
	return nil
}

func (c *systemstatisticsUDPCollector) collectSystemStatisticsUDP(ch chan<- prometheus.Metric, labelValues []string, s StatisticsUDP) {
	l := append(labelValues, "udp")
	ch <- prometheus.MustNewConstMetric(datagramsReceivedDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsReceived, l...)
	ch <- prometheus.MustNewConstMetric(datagramsWithIncompleteHeaderDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsWithIncompleteHeader, l...)
	ch <- prometheus.MustNewConstMetric(datagramsWithBadDatalengthFieldDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsWithBadDatalengthField, l...)
	ch <- prometheus.MustNewConstMetric(datagramsWithBadChecksumDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsWithBadChecksum, l...)
	ch <- prometheus.MustNewConstMetric(datagramsDroppedDueToNoSocketDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsDroppedDueToNoSocket, l...)
	ch <- prometheus.MustNewConstMetric(broadcastOrMulticastDatagramsDroppedDueToNoSocketDesc, prometheus.CounterValue, s.Statistics.Udp.BroadcastOrMulticastDatagramsDroppedDueToNoSocket, l...)
	ch <- prometheus.MustNewConstMetric(datagramsDroppedDueToFullSocketBuffersDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsDroppedDueToFullSocketBuffers, l...)
	ch <- prometheus.MustNewConstMetric(datagramsNotForHashedPcbDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsNotForHashedPcb, l...)
	ch <- prometheus.MustNewConstMetric(datagramsDeliveredDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsDelivered, l...)
	ch <- prometheus.MustNewConstMetric(datagramsOutputDesc, prometheus.CounterValue, s.Statistics.Udp.DatagramsOutput, l...)
}
