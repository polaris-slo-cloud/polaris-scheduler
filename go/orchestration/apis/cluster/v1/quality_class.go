package v1

// NetworkQualityClass is used to describe the advertised performance of a NetworkLink
//
// +kubebuilder:validation:Enum=QC1Mbps;QC10Mbps;QC100Mbps;QC1Gbps;QC10Gbps
type NetworkQualityClass string

const (
	QC1Mbps   NetworkQualityClass = "QC1Mbps"
	QC10Mbps  NetworkQualityClass = "QC10Mbps"
	QC100Mbps NetworkQualityClass = "QC100Mbps"
	QC1Gbps   NetworkQualityClass = "QC1Gbps"
	QC10Gbps  NetworkQualityClass = "QC10Gbps"
)

var (
	qualityClassToKbps map[NetworkQualityClass]int64 = map[NetworkQualityClass]int64{
		QC1Mbps:   1000,
		QC10Mbps:  10000,
		QC100Mbps: 100000,
		QC1Gbps:   1000000,
		QC10Gbps:  10000000,
	}
)

// Converts a NetworkQualityClass enum value to its equivalent int Kbps.
func NetworkQualitClassToKbps(qualityClass NetworkQualityClass) int64 {
	return qualityClassToKbps[qualityClass]
}
