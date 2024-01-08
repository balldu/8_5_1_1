package plugins

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"math"
	"math/rand"
	//"k8s.io/apimachinery/pkg/api/resource"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	//"math/rand"
	//"time"
)

const F_Name = "custom-plugin"

type Args_ATH struct {
	// 限制Node可运行的实例数量，初始化为3，增大，步幅为1
	alpha int
	// 限制Node的最大资源余量与最小资源余量的差值，增大，步幅为0.05百分比，此处均为该类型资源的百分比进行运算
	beta float64
	// 设定比较近期租户的数量定值比较
	W int
	// 阈值，在Node近期安置的租户达到定值W时。进行比较
	gamma float64
	// 选择目标服务器，这边表示为节点
	m_star int
	// 应用实例a的资源开销, 本身的CPU开销分别设置为0.045和0.015
	C_application []float64
	// 目前的用户数
	Num int
	// 最差SLA性能要求的单位用户使用应用时资源的开销，称为应用对资源的基准开销,q{1,2,3}, a =1,...N
	f_a_q [][]float64
	// 矩阵A ，如果服务器上有应用的实例，则为1，否则为0
	A_a_m [][]int
	// 用于记录服务器有了哪些pod
	record []int
	// 节点数为{5, 10, 15}
	M int
	// 资源类型数：3，包含：CPU,内存,硬盘
	p int
	// 应用数: {5, 10, 15}
	N int
	// 服务器每种资源的总量
	m_total_resources [][]float64
	// 应用对资源的开销（随机生成）
	a_r [][]float64
	// 租户数
	tenants int
	// 租户的活跃用户数（随机生成）5-50
	tenant []int
	// 租户请求的应用t_a 0-N
	t_a []int
	// 给服务器设置参数w_m计数服务器上近期安置的租户
	w_m int
	// 选择的目标应用
	a_star int
}

var a Args_ATH

// 初始化
func (a *Args_ATH) init() {
	// 限制服务器可运行的实例数量，初始化为3，增大，步幅为1
	a.alpha = 3
	// 限制服务器的最大资源余量与最小资源余量的差值，增大，步幅为0.05百分比，此处均为该类型资源的百分比进行运算
	a.beta = 0.9
	// 设定比较近期租户的数量定值比较
	a.W = 150
	// 阈值，在服务器近期安置的租户达到定值W时。进行比较
	a.gamma = 0.9
	// 选择目标服务器，这边表示为节点
	a.m_star = -1
	// 应用实例a的资源开销, 本身的CPU开销分别设置为0.045和0.015
	a.C_application = make([]float64, 10)
	// 目前的用户数
	a.Num = 0
	// 最差SLA性能要求的单位用户使用应用时资源的开销，称为应用对资源的基准开销,q{1,2,3}, a =1,...N
	// 随机生成
	a.f_a_q = make([][]float64, 15)
	for i := range a.f_a_q {
		a.f_a_q[i] = make([]float64, 3)
	}
	for i := range a.a_r {
		for j := range a.a_r[i] {
			a.a_r[i][j] = 0.31 + 0.31*rand.Float64()
		}
	}
	// 矩阵A ，如果服务器上有应用的实例，则为1，否则为0
	a.A_a_m = make([][]int, 10)
	for i := range a.A_a_m {
		a.A_a_m[i] = make([]int, 5)
	}
	// 用于记录服务器有了哪些应用
	a.record = make([]int, 15)
	// 节点数为{5, 10, 15}
	a.M = 5
	// 资源类型数：3，包含：CPU,内存,硬盘
	a.p = 3
	// 应用数: {5, 10, 15}
	a.N = 10
	// 服务器的资源总量
	a.m_total_resources = make([][]float64, 5)
	for j := range a.m_total_resources {
		a.m_total_resources[j] = make([]float64, 3)
		// 设置其中一种的CPU本身开销为0.045
		// 主节点的初始化
		a.m_total_resources[0] = []float64{1 - 0.045, 0.74, 0.168}
		// 节点1 的初始化
		a.m_total_resources[1] = []float64{1 - 0.045, 0.83, 0.808}
		// 节点2 的初始化
		a.m_total_resources[2] = []float64{1 - 0.045, 0.83, 0.811}
	}
	// 应用对资源的开销，（随机生成）, 按照百分比
	a.a_r = make([][]float64, 10)
	for i := range a.a_r {
		a.a_r[i] = make([]float64, 3)
	}
	for i := range a.a_r {
		for j := range a.a_r[i] {
			a.a_r[i][j] = 0.01 + 0.04*rand.Float64()
		}
	}
	// 租户数
	a.tenants = 20000
	// 租户的活跃用户数（随机生成）5-50
	a.tenant = make([]int, a.tenants)
	for i := 0; i < a.tenants; i++ {
		a.tenant[i] = 5 + rand.Intn(45)
	}
	// 租户请求的应用t_a 0-N
	a.t_a = make([]int, a.tenants)
	for i := 0; i < a.tenants; i++ {
		a.t_a[i] = rand.Intn(a.N)
	}
	// 给服务器设置参数w_m计数服务器上近期安置的租户
	a.w_m = 0
	// 选择的目标应用
	a.a_star = -1

}

func (s *Sample) F_Name() string {
	return F_Name
}

func (s *Sample) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// Filter 回调方法
func (sr Sample) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	// 在这里实现过滤逻辑
	fmt.Println("进入PreFilter...............")
	fmt.Printf("pod:%s,namespace:%s\n", pod.Name, pod.Namespace)
	a.init()
	fmt.Println("-----------------------------------------")
	return framework.NewStatus(framework.Success) // Pod 可以在此节点上调度
}

func (sr Sample) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeinfo *nodeinfo.NodeInfo) *framework.Status {
	fmt.Println("进入Filter...............")
	fmt.Printf("Pod: %s, Namespace: %s\n", pod.Name, pod.Namespace)
	nodeName := nodeinfo.Node().Name
	fmt.Printf("NodeName: %v\n", nodeName)
	fmt.Println("-----------------------------------------")

	// 统计 Pod 请求的资源量
	var totalRequestedCpu, totalRequestedMemory, totalRequestedStorage int64
	for _, container := range pod.Spec.Containers {
		totalRequestedCpu += container.Resources.Requests.Cpu().MilliValue()
		totalRequestedMemory += container.Resources.Requests.Memory().Value()
		totalRequestedStorage += container.Resources.Requests.StorageEphemeral().Value()
	}

	// 获取节点的可用资源
	nodeInfo, err := sr.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error(), "无法获得节点的信息")
	}
	nodeResources := nodeInfo.AllocatableResource()
	nodeAvailCpu := nodeResources.MilliCPU
	nodeAvailMemory := nodeResources.Memory
	nodeAvailStorage := nodeResources.EphemeralStorage
	max_per := math.Max(math.Max(float64(nodeAvailCpu)/2000.0, float64(nodeAvailMemory)/(20*1024*1024*1024)),
		float64(nodeAvailStorage)/(20*1024*1024*1024))
	min_per := math.Min(math.Min(float64(nodeAvailCpu)/2000, float64(nodeAvailMemory)/(20*1024*1024*1024)),
		float64(nodeAvailStorage)/(20*1024*1024*1024))
	fmt.Printf("nodeAvailCpu: %v\n", float64(nodeAvailCpu)/2000.0)
	fmt.Printf("nodeAvailMemory: %v\n", float64(nodeAvailMemory)/(20*1024*1024*1024))
	fmt.Printf("nodeAvailStorage: %v\n", float64(nodeAvailStorage)/(20*1024*1024*1024))
	fmt.Printf("max_per: %v\n", max_per)
	fmt.Printf("min_per: %v\n", min_per)
	// 对比节点资源和 Pod 请求资源
	if nodeAvailCpu < totalRequestedCpu {
		return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("节点 %s CPU 不足", nodeInfo.Node().Name))
	}
	if nodeAvailMemory < totalRequestedMemory {
		return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("节点 %s 内存不足", nodeInfo.Node().Name))
	}
	if nodeAvailStorage < totalRequestedStorage {
		return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("节点 %s 存储空间不足", nodeInfo.Node().Name))
	}
	if max_per-min_per > a.beta {
		return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("节点 %s 最大资源余量与最小资源余量超过beta %v", nodeInfo.Node().Name, a.beta))
	}
	// 所有检查都通过了，表示 Pod 可以在该节点上调度
	return framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 可以调度"))
}

// type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
func New_M(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(configuration, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)
	return &Sample{
		args:   args,
		handle: f,
	}, nil
}
