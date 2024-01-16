package plugins

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

type PredictedResources struct {
	gpuCores  int
	gpumem    int
	bandwidth int
}

// Clone returns a deep copy of PredictedResources.
func (pr *PredictedResources) Clone() framework.StateData {
	return &PredictedResources{
		gpuCores:  pr.gpuCores,
		gpumem:    pr.gpumem,
		bandwidth: pr.bandwidth,
	}
}

// 转换注解中字符串为整型
func parseAnnotation(pod *v1.Pod, key string) (int, error) {
	if value, ok := pod.Annotations[key]; ok {
		return strconv.Atoi(value)
	}
	return 0, fmt.Errorf("annotation %s not found", key)
}

func (sr *Sample) predictResources(cycleState *framework.CycleState, pod *v1.Pod) error {
	// 从 Pod 的注解中提取 featureData
	batchSize, _ := parseAnnotation(pod, "batch_size")
	numBatches, _ := parseAnnotation(pod, "num_batches")
	numPS, _ := parseAnnotation(pod, "num_ps")
	numWorker, _ := parseAnnotation(pod, "num_worker")
	fmt.Printf("Feature input : batchSize: %d, numBatches: %d, numPS: %d, numWorker: %d\n", batchSize, numBatches, numPS, numWorker)
	// 用于模型预测的数据结构（一个 slice of floats）
	featureData := []float64{float64(batchSize), float64(numBatches), float64(numPS), float64(numWorker)}
	featureData2 := [][]float64{featureData}
	// 使用模型进行预测
	pred1 := sr.model1.Predict(featureData2)
	pred2 := sr.model2.Predict(featureData2)
	pred3 := sr.model3.Predict(featureData2)
	predictedResources := PredictedResources{
		gpuCores:  int(pred1[0]),
		gpumem:    int(pred2[0]),
		bandwidth: int(pred3[0]),
	}
	// 打印预测结果
	fmt.Printf("Predicted resources - GPU Cores: %d, GPU Memory: %d, Bandwidth: %d\n",
		predictedResources.gpuCores,
		predictedResources.gpumem,
		predictedResources.bandwidth,
	)
	// Store the predicted resources in CycleState
	cycleState.Write(PredictedResourcesKey, &predictedResources)

	return nil
}

// This is an example within a Filter or Score plugin method
func getPredictedResources(cycleState *framework.CycleState) (*PredictedResources, error) {
	c, err := cycleState.Read(PredictedResourcesKey)
	if err != nil {
		// no prediction data available, handle error appropriately
		return nil, fmt.Errorf("error reading %s from cycleState: %v", PredictedResourcesKey, err)
	}

	predictedResources, ok := c.(*PredictedResources)
	if !ok {
		// the stored data is not of the expected type
		return nil, fmt.Errorf("expected *PredictedResources type, but got %T", c)
	}

	return predictedResources, nil
}

func (sr Sample) getScore(state *framework.CycleState, nodeName string) int64 {
	fmt.Println("进入getRFScore........")
	// 获取节点的信息
	nodeInfo, err := sr.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		// 处理错误，例如打印日志、返回错误等
		fmt.Println("获取节点信息失败:", err)
		return 100
	}
	fmt.Println(nodeInfo.String())
	resource := nodeInfo.AllocatableResource() // 调用函数获取 nodeinfo.Resource 实例
	milliCPU := resource.MilliCPU              // 访问 MilliCPU 字段
	memory := resource.Memory                  // 访问 Memory 字段
	fmt.Printf("MilliCPU: %d, 类型: %T\n", milliCPU, milliCPU)
	fmt.Printf("Memory: %d, 类型: %T\n", memory, memory)

	predictedResources, err := getPredictedResources(state)
	cpuScore := milliCPU / int64(predictedResources.gpuCores)
	memoryScore := memory / (int64(predictedResources.gpumem) * 1024)
	score := (cpuScore + memoryScore) / 2
	fmt.Println("score of node %s: %d", nodeName, score)
	return score
}

var featureData = [][]float64{
	[]float64{32, 100, 1, 4},
	[]float64{16, 50, 2, 8},
	[]float64{64, 200, 4, 16},
	[]float64{128, 500, 8, 32},
	[]float64{256, 1000, 16, 64},
}

var Num_GPU = []float64{
	1.0,  // 对应可能的 GPU 个数
	2.0,  // 对应可能的 GPU 个数
	4.0,  // 对应可能的 GPU 个数
	8.0,  // 对应可能的 GPU 个数
	16.0, // 对应可能的 GPU 个数
}

var GMem = []float64{
	4.0,  // 对应可能的 GPU 内存大小，以 GB 为单位
	8.0,  // 对应可能的 GPU 内存大小，以 GB 为单位
	16.0, // 对应可能的 GPU 内存大小，以 GB 为单位
	32.0, // 对应可能的 GPU 内存大小，以 GB 为单位
	64.0, // 对应可能的 GPU 内存大小，以 GB 为单位
}

var Bandwidth = []float64{
	80.0,  // 对应可能的带宽大小，以 GB/s 为单位
	120.0, // 对应可能的带宽大小，以 GB/s 为单位
	160.0, // 对应可能的带宽大小，以 GB/s 为单位
	240.0, // 对应可能的带宽大小，以 GB/s 为单位
	320.0, // 对应可能的带宽大小，以 GB/s 为单位
}
