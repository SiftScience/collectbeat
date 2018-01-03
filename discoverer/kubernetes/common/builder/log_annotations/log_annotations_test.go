package log_annotations

import (
	"encoding/json"
	"testing"

	"github.com/ebay/collectbeat/discoverer/common/builder"

	"github.com/elastic/beats/libbeat/common"
	kubernetes "github.com/elastic/beats/libbeat/processors/add_kubernetes_metadata"

	"github.com/stretchr/testify/assert"
)

func TestLogAnnotationBuilder(t *testing.T) {
	b, ok := getLogAnnotationBuilder(t)
	assert.Equal(t, ok, true)

	tests := []struct {
		annotations map[string]interface{}
		length      int
	}{
		{
			annotations: map[string]interface{}{},
			length:      2,
		},
		{
			annotations: map[string]interface{}{
				"foo/pattern": "bar",
			},
			length: 2,
		},
		{
			annotations: map[string]interface{}{
				"foo.nginx/pattern": "abc",
			},
			length: 2,
		},
		{
			annotations: map[string]interface{}{
				"foo.nginx/pattern":  "abc",
				"foo.apache/pattern": "cde",
			},
			length: 2,
		},
		{
			annotations: map[string]interface{}{
				"foo/logzCodec": "json",
				"foo/logzEnv":   "prod",
			},
			length: 2,
		},
		{
			annotations: map[string]interface{}{
				"foo/pattern":   "bar",
				"foo/logzCodec": "json",
				"foo/logzEnv":   "prod",
			},
			length: 2,
		},
	}

	for _, test := range tests {
		iface := map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace":   "foo",
				"name":        "bar",
				"annotations": test.annotations,
			},
			"status": map[string]interface{}{
				"podIP": "4.5.6.7",
				"containerStatuses": []map[string]interface{}{
					{
						"containerID": "docker://123",
						"name":        "nginx",
					},
					{
						"containerID": "docker://456",
						"name":        "apache",
					},
				},
			},
		}
		pod := &kubernetes.Pod{}

		data, _ := json.Marshal(iface)
		json.Unmarshal(data, pod)

		confs := b.BuildModuleConfigs(pod)
		ok := assert.Equal(t, len(confs), 2)
		if !ok {
			t.FailNow()
		}
	}
}
func getLogAnnotationBuilder(t *testing.T) (builder.PollerBuilder, bool) {
	cfg := map[string]interface{}{
		"prefix":            "foo",
		"default_namespace": "abc",
		"logs_path":         "/var/",
	}
	config, _ := common.NewConfigFrom(cfg)
	bRaw, err := NewPodLogAnnotationBuilder(config, nil, nil)
	assert.NotNil(t, bRaw)
	assert.Nil(t, err)
	b, ok := bRaw.(builder.PollerBuilder)
	return b, ok
}

func TestProspectorConfig(t *testing.T) {
	b, ok := getLogAnnotationBuilder(t)
	assert.Equal(t, ok, true)

	iface := map[string]interface{}{
		"metadata": map[string]interface{}{
			"namespace": "foo",
			"name":      "bar",
			"annotations": map[string]interface{}{
				"foo.nginx/pattern":  "abc",
				"foo.apache/pattern": "cde",
			},
		},
		"status": map[string]interface{}{
			"podIP": "4.5.6.7",
			"containerStatuses": []map[string]interface{}{
				{
					"containerID": "docker://123",
					"name":        "nginx",
				},
				{
					"containerID": "docker://456",
					"name":        "apache",
				},
			},
		},
	}
	pod := &kubernetes.Pod{}

	data, _ := json.Marshal(iface)
	json.Unmarshal(data, pod)

	confs := b.BuildModuleConfigs(pod)
	ok = assert.Equal(t, len(confs), 2)
	if !ok {
		t.FailNow()
	}

	multilineCfg := common.MapStr{}
	setMultilineConfig(multilineCfg, "abc", false, "after")

	assert.Equal(t, confs[0].Config["paths"], []string{"/var/123/*.log"})
	assert.Equal(t, confs[0].Config["multiline"], multilineCfg["multiline"])

	setMultilineConfig(multilineCfg, "cde", false, "after")
	assert.Equal(t, confs[1].Config["paths"], []string{"/var/456/*.log"})
	assert.Equal(t, confs[1].Config["multiline"], multilineCfg["multiline"])

	logzCfg := common.MapStr{}
	logzExpectedCfg := common.MapStr{
		"logzToken": "ABC123",
		"logzCodec": "json",
		"logzEnv":   "prod",
	}
	setLogzFields("ABC123", "json", "prod", logzCfg)
	assert.Equal(t, logzExpectedCfg, logzCfg["fields"])

	logzCfg["fields"] = common.MapStr{
		"namespace": "abc",
	}
	logzExpectedCfg["namespace"] = "abc"
	setLogzFields("ABC123", "json", "prod", logzCfg)
	assert.Equal(t, logzExpectedCfg, logzCfg["fields"])
}
