package report

import (
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/elasticsearch/pkg/controller"
	"github.com/kubedb/elasticsearch/pkg/util/es"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ExportReport(
	kubeClient kubernetes.Interface,
	dbClient cs.Interface,
	namespace string,
	kubedbName string,
	index string,
	w http.ResponseWriter,
) {
	startTime := metav1.Now()

	elastic, err := dbClient.KubedbV1alpha1().Elasticsearches(namespace).Get(kubedbName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			http.Error(w, fmt.Sprintf(`Elasticsearch "%v" not found`, kubedbName), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	url := elastic.GetConnectionURL()
	c := controller.New(nil, kubeClient, nil, nil, nil, nil, nil, nil, amc.Config{}, nil)
	client, err := es.GetElasticClient(c.Client, c.ExtClient, elastic, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Stop()
	indices := make([]string, 0)
	if index == "" {
		indices, err = client.GetIndexNames()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		indices = append(indices, index)
	}

	esSummary := make(map[string]*api.ElasticsearchSummary)
	for _, index := range indices {
		info, err := client.GetElasticsearchSummary(index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		esSummary[index] = info
	}

	completionTime := metav1.Now()

	r := &api.Report{
		TypeMeta:   elastic.TypeMeta,
		ObjectMeta: elastic.ObjectMeta,
		Summary: api.ReportSummary{
			Elasticsearch: esSummary,
		},
		Status: api.ReportStatus{
			StartTime:      &startTime,
			CompletionTime: &completionTime,
		},
	}
	r.ResourceVersion = ""
	r.SelfLink = ""
	r.UID = ""

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(data))
	} else {
		http.Error(w, "audit data not found", http.StatusNotFound)
	}
}
