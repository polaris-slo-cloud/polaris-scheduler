package v1

// NetworkQualityClass is used to describe the advertised performance of a NetworkLink
//
// +kubebuilder:validation:Enum=QC1Mbps;QC10Mbps;QC1Gbps
type NetworkQualityClass string

// ToDo: Add more constants.
const (
	QC1Mbps  NetworkQualityClass = "QC1Mbps"
	QC10Mbps NetworkQualityClass = "QC10Mbps"
	QC1Gbps  NetworkQualityClass = "QC1Gbps"
)
