package main

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	//"fmt"
	"html/template"
	"log"
	"net/http"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodStats struct {
	Namespace string
	PodName   string
	CPUUsage  string
	MemUsage  string
}

type PodStatsData struct {
	Stats []PodStats
}

func main() {
	// 初始化Kubernetes客户端
	config, err := clientcmd.BuildConfigFromFlags("", "/root/GetPodResource/k3s.yaml")
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// 初始化Metrics客户端
	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// 获取Pod资源开销统计数据

	stats, err := getPodStats(clientset, metricsClient)
	if err != nil {
		log.Fatal(err)
	}

	// 注册HTTP处理函数
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 渲染模板并将数据传递给模板
		tmpl, err := template.ParseFiles("template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, stats)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 启动Web服务器
	fmt.Printf("Server running on http://localhost:8080")
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// 获取Pod资源开销统计数据
func getPodStats(clientset kubernetes.Interface, metricsClient metricsv.Interface) (PodStatsData, error) {
	stats := PodStatsData{}
	podListOptions := v1.ListOptions{
		FieldSelector: "status.phase=Running",
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), podListOptions)
	//pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return stats, err
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running"{

		} else {
			fmt.Printf("9999999999999")
		}
		namespace := pod.Namespace
		podName := pod.Name

		podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(context.TODO(), podName, v1.GetOptions{})
		if err != nil {
			return stats, err
		}

		for _, container := range podMetrics.Containers {
			cpuUsage := container.Usage["cpu"]
			memUsage := container.Usage["memory"]
			fmt.Printf("6666, %v:%s\n", cpuUsage, reflect.TypeOf(cpuUsage))
			fmt.Printf("6666, %s\n", reflect.TypeOf(memUsage))

			//cpuUsageStr := cpuUsage.String()
			//memUsageStr := memUsage.String()
			cpuUsageStr := cpuFormatBytes(cpuUsage.String())
			memUsageStr := memFormatBytes(memUsage.String())


			stat := PodStats{
				Namespace: namespace,
				PodName:   podName,
				CPUUsage:  cpuUsageStr,
				MemUsage:  memUsageStr,
			}
			stats.Stats = append(stats.Stats, stat)
		}
	}

	return stats, nil
}


func cpuFormatBytes(bytes string) string {
	res := removeCharacter(bytes, "n")
	f, err := strconv.ParseFloat(res, 64)
	if err != nil {
		fmt.Println("Failed to convert string to float64:", err)
	}

	result := f / 1000 / 1000
	str := strconv.FormatFloat(result, 'f', 2, 64)
	return str + "m"
}

func memFormatBytes(bytes string) string {
	res := removeCharacter(bytes, "Ki")
	f, err := strconv.ParseFloat(res, 64)
	if err != nil {
		fmt.Println("Failed to convert string to float64:", err)
	}

	result := f / 1024
	str := strconv.FormatFloat(result, 'f', 2, 64)
	return str + "Mi"
}


func removeCharacter(str, char string) string {
	result := strings.ReplaceAll(str, char, "")
	return result
}
