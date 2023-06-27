package prettyprint

import (
	v1 "github.com/JoeReid/jetbridge/proto/gen/go/jetbridge/v1"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func Binding(binding *v1.JetstreamBinding) {
	Bindings([]*v1.JetstreamBinding{binding})
}

func Bindings(bindings []*v1.JetstreamBinding) {
	tbl := table.New("ID", "Lambda ARN", "Stream", "Subject", "Max Messages", "Max Latency", "Assigned Peer")

	tbl.WithHeaderFormatter(color.New(color.FgGreen, color.Underline).SprintfFunc())
	tbl.WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc())

	for _, binding := range bindings {
		vals := []interface{}{
			binding.Id,
			binding.LambdaArn,
			binding.Stream,
			binding.SubjectPattern,
		}

		if binding.MaxBatchSize == 0 {
			vals = append(vals, "-")
		} else {
			vals = append(vals, binding.MaxBatchSize)
		}

		if binding.MaxBatchLatency.AsDuration() == 0 {
			vals = append(vals, "-")
		} else {
			vals = append(vals, binding.MaxBatchLatency.AsDuration())
		}

		vals = append(vals, binding.AssignedPeer)

		tbl.AddRow(vals...)
	}
	tbl.Print()
}

func Peers(peers []*v1.Peer) {
	tbl := table.New("ID", "Hostname", "Joined At", "Last Seen", "Heartbeat Due By")

	tbl.WithHeaderFormatter(color.New(color.FgGreen, color.Underline).SprintfFunc())
	tbl.WithFirstColumnFormatter(color.New(color.FgYellow).SprintfFunc())

	for _, peer := range peers {
		tbl.AddRow(peer.Id, peer.Hostname, peer.Joined.AsTime(), peer.LastSeen.AsTime(), peer.HeartbeatDue.AsTime())
	}
	tbl.Print()
}
