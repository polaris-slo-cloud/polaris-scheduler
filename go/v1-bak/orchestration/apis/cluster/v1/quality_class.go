package v1

// NetworkQualityClass is used to describe the advertised performance of a NetworkLink
//
// +kubebuilder:validation:Enum=QC1Mbps;QC2Mbps;QC3Mbps;QC4Mbps;QC5Mbps;QC6Mbps;QC7Mbps;QC8Mbps;QC9Mbps;QC10Mbps;QC20Mbps;QC30Mbps;QC40Mbps;QC50Mbps;QC60Mbps;QC70Mbps;QC80Mbps;QC90Mbps;QC100Mbps;QC1Gbps;QC2Gbps;QC3Gbps;QC4Gbps;QC5Gbps;QC6Gbps;QC7Gbps;QC8Gbps;QC9Gbps;QC10Gbps
type NetworkQualityClass string

const (
	QC1Mbps   NetworkQualityClass = "QC1Mbps"
	QC2Mbps   NetworkQualityClass = "QC2Mbps"
	QC3Mbps   NetworkQualityClass = "QC3Mbps"
	QC4Mbps   NetworkQualityClass = "QC4Mbps"
	QC5Mbps   NetworkQualityClass = "QC5Mbps"
	QC6Mbps   NetworkQualityClass = "QC6Mbps"
	QC7Mbps   NetworkQualityClass = "QC7Mbps"
	QC8Mbps   NetworkQualityClass = "QC8Mbps"
	QC9Mbps   NetworkQualityClass = "QC9Mbps"
	QC10Mbps  NetworkQualityClass = "QC10Mbps"
	QC20Mbps  NetworkQualityClass = "QC20Mbps"
	QC30Mbps  NetworkQualityClass = "QC30Mbps"
	QC40Mbps  NetworkQualityClass = "QC40Mbps"
	QC50Mbps  NetworkQualityClass = "QC50Mbps"
	QC60Mbps  NetworkQualityClass = "QC60Mbps"
	QC70Mbps  NetworkQualityClass = "QC70Mbps"
	QC80Mbps  NetworkQualityClass = "QC80Mbps"
	QC90Mbps  NetworkQualityClass = "QC90Mbps"
	QC100Mbps NetworkQualityClass = "QC100Mbps"
	QC1Gbps   NetworkQualityClass = "QC1Gbps"
	QC2Gbps   NetworkQualityClass = "QC2Gbps"
	QC3Gbps   NetworkQualityClass = "QC3Gbps"
	QC4Gbps   NetworkQualityClass = "QC4Gbps"
	QC5Gbps   NetworkQualityClass = "QC5Gbps"
	QC6Gbps   NetworkQualityClass = "QC6Gbps"
	QC7Gbps   NetworkQualityClass = "QC7Gbps"
	QC8Gbps   NetworkQualityClass = "QC8Gbps"
	QC9Gbps   NetworkQualityClass = "QC9Gbps"
	QC10Gbps  NetworkQualityClass = "QC10Gbps"
)

var (
	qualityClassToKbps map[NetworkQualityClass]int64 = map[NetworkQualityClass]int64{
		QC1Mbps:   1000,
		QC2Mbps:   2000,
		QC3Mbps:   3000,
		QC4Mbps:   4000,
		QC5Mbps:   5000,
		QC6Mbps:   6000,
		QC7Mbps:   7000,
		QC8Mbps:   8000,
		QC9Mbps:   9000,
		QC10Mbps:  10000,
		QC20Mbps:  20000,
		QC30Mbps:  30000,
		QC40Mbps:  40000,
		QC50Mbps:  50000,
		QC60Mbps:  60000,
		QC70Mbps:  70000,
		QC80Mbps:  80000,
		QC90Mbps:  90000,
		QC100Mbps: 100000,
		QC1Gbps:   1000000,
		QC2Gbps:   2000000,
		QC3Gbps:   3000000,
		QC4Gbps:   4000000,
		QC5Gbps:   5000000,
		QC6Gbps:   6000000,
		QC7Gbps:   7000000,
		QC8Gbps:   8000000,
		QC9Gbps:   9000000,
		QC10Gbps:  10000000,
	}
)

// Converts a NetworkQualityClass enum value to its equivalent int Kbps.
func NetworkQualitClassToKbps(qualityClass NetworkQualityClass) int64 {
	return qualityClassToKbps[qualityClass]
}
