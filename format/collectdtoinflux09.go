// Copyright 2015-2016 trivago GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package format

import (
	"fmt"
	"github.com/trivago/gollum/core"
	"github.com/trivago/tgo/tio"
	"github.com/trivago/tgo/tmath"
)

// CollectdToInflux09 provides a transformation from collectd JSON data to
// InfluxDB 0.9.x compatible JSON data. Trailing and leading commas are removed
// from the Collectd message beforehand.
// Configuration example
//
//   - "<producer|stream>":
//     Formatter: "format.CollectdToInflux09"
//     CollectdToInfluxFormatter: "format.Forward"
//
// CollectdToInfluxFormatter defines the formatter applied before the conversion
// from Collectd to InfluxDB. By default this is set to format.Forward.
type CollectdToInflux09 struct {
	core.FormatterBase
}

func init() {
	core.TypeRegistry.Register(CollectdToInflux09{})
}

// Configure initializes this formatter with values from a plugin config.
func (format *CollectdToInflux09) Configure(conf core.PluginConfig) error {
	return format.FormatterBase.Configure(conf)
}

// Format transforms collectd data to influx 0.9.x data
func (format *CollectdToInflux09) Format(msg core.Message) ([]byte, core.MessageStreamID) {
	collectdData, err := parseCollectdPacket(msg.Data)
	if err != nil {
		format.Log.Error.Print("Collectd parser error: ", err)
		return []byte{}, msg.StreamID // ### return, error ###
	}

	// Manually convert to JSON lines
	influxData := tio.NewByteStream(len(msg.Data))
	fixedPart := fmt.Sprintf(
		`{"name": "%s", "timestamp": %d, "precision": "ms", "tags": {"plugin_instance": "%s", "type": "%s", "type_instance": "%s", "host": "%s"`,
		collectdData.Plugin,
		int64(collectdData.Time),
		collectdData.PluginInstance,
		collectdData.PluginType,
		collectdData.TypeInstance,
		collectdData.Host)

	setSize := tmath.Min3I(len(collectdData.Dstypes), len(collectdData.Dsnames), len(collectdData.Values))
	for i := 0; i < setSize; i++ {
		fmt.Fprintf(&influxData,
			`%s, "dstype": "%s", "dsname": "%s"}, "fields": {"value": %f} },`,
			fixedPart,
			collectdData.Dstypes[i],
			collectdData.Dsnames[i],
			collectdData.Values[i])
	}

	return influxData.Bytes(), msg.StreamID
}
