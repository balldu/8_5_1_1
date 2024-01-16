package plugins

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// 插件名称
const Name = "custom-plugin"

// 表征pod的下标
var t int64 = 0

// 表征当前最高分数的节点
var nodename string = ""

type Args struct {
	FavoriteColor  string `json:"favorite_color,omitempty"`
	FavoriteNumber int    `json:"favorite_number,omitempty"`
	ThanksTo       string `json:"thanks_to,omitempty"`
}

type Sample struct {
	args   *Args
	handle framework.FrameworkHandle
}

func (s *Sample) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	fmt.Println("进入NormalizeScore........")
	fmt.Printf("pod:%s,namespace:%s\n", pod.Name, pod.Namespace)
	fmt.Println("-----------------------------------------")
	var maxScore int64 = 0
	for _, ns := range scores {
		if ns.Score > maxScore {
			maxScore = ns.Score
		}
	}
	if maxScore == 0 {
		return framework.NewStatus(framework.Success)
	}
	// 进行归一化处理
	for i := range scores {
		scores[i].Score = scores[i].Score * framework.MaxNodeScore / maxScore
		fmt.Printf("the %v of scores is %v\n", i, scores[i].Score)
	}
	nodeName := s.GetMaxScoreNode(scores)
	nodename = nodeName
	return framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 符合标准，并且分数最高！！", nodeName))
}

func (s *Sample) Score_Extended() (int64, *framework.Status) {
	// C_t + C_a
	C_t := make([]float64, 3)
	C_t = []float64{float64(a.tenant[t]) * a.a_r[t_a][0], float64(a.tenant[t]) * a.a_r[t_a][1], float64(a.tenant[t]) * a.a_r[t_a][0]}
	C_sum := make([]float64, len(C_t))
	for i := range C_sum {
		C_sum[i] = C_t[i] + a.C_application[i]
	}
	Res_m_2 := a.Compute_VD(C_sum)
	min := Res_m_2[0]
	for i := 0; i < a.M; i++ {
		Records := a.ATH()
		if Res_m_2[i] <= min && a.A_a_m[t_a][i] == 0 && (a.record[Records[t_a][1]]+a.record[Records[t_a][0]]+a.record[Records[t_a][2]] >= 1) && a.Num < a.alpha {
			a.m_star = i
			min = Res_m_2[i]
		}
	}
	if a.m_star != -1 {
		a.m_total_resources[a.m_star][0] = a.m_total_resources[a.m_star][0] - C_t[0]
		a.m_total_resources[a.m_star][1] = a.m_total_resources[a.m_star][1] - C_t[0]
		a.m_total_resources[a.m_star][2] = a.m_total_resources[a.m_star][2] - C_t[0]
		a.Num += 1
		a.w_m += 1
	} else {
		min = Res_m_2[0]
		for i := 0; i < a.M; i++ {
			Records := a.ATH()
			if Res_m_2[i] <= min && a.A_a_m[t_a][i] == 0 && (a.record[Records[t_a][1]]+a.record[Records[t_a][0]]+a.record[Records[t_a][2]] < 1) && a.Num < a.alpha {
				a.m_star = i
				min = Res_m_2[i]
			}
			if a.m_star != -1 {
				a.m_total_resources[a.m_star][0] = a.m_total_resources[a.m_star][0] - C_t[0]
				a.m_total_resources[a.m_star][1] = a.m_total_resources[a.m_star][1] - C_t[0]
				a.m_total_resources[a.m_star][2] = a.m_total_resources[a.m_star][2] - C_t[0]
				a.Num += 1
				a.w_m += 1
				a.tenants -= 1
			}
		}
	}
	if a.beta < 1 {
		a.beta += 0.05
	} else if a.beta == 1 && a.alpha < a.N {
		a.beta = 0.2
	} else if a.beta == 1 && a.alpha == a.N {
		return -1, framework.NewStatus(framework.Unschedulable, fmt.Sprintf("目前没有找到合适的节点"))
	}
	if a.m_star == 0 {
		// 10表示直接就能够找到合适的服务器节点
		return 10, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-master"))
	} else if a.m_star == 1 {
		return 10, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-worker1"))
	} else if a.m_star == 2 {
		return 10, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-worker2"))
	}
	return 1, framework.NewStatus(framework.Unschedulable, fmt.Sprintf("目前没有找到合适的节点"))
}

func (s *Sample) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *nodeinfo.NodeInfo) *framework.Status {
	fmt.Println("进入AddPod...............")
	fmt.Println("-----------------------------------------")
	return framework.NewStatus(framework.Success)
}

func (s *Sample) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podToRemove *v1.Pod, nodeInfo *nodeinfo.NodeInfo) *framework.Status {
	fmt.Println("进入RemovePod...............")
	fmt.Println("-----------------------------------------")
	return framework.NewStatus(framework.Success)
}

func (s *Sample) Name() string {
	return Name
}

func (sr Sample) PreBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	fmt.Println("进入PreBind........")
	fmt.Printf("pod:%s,namespace:%s\n", pod.Name, pod.Namespace)
	fmt.Printf("nodeName:%v\n", nodeName)
	fmt.Println("-----------------------------------------")
	if pod == nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("pod cannot be nil"))
	}
	return nil
}

func (sr Sample) GetMaxScoreNode(scores framework.NodeScoreList) string {
	var maxScore int64 = 0
	var nodeName string
	var node_index int
	for node, score := range scores {
		if score.Score > maxScore {
			maxScore = score.Score
			node_index = node
		}
	}
	if node_index == 0 {
		nodeName = "master"
	} else if node_index == 1 {
		nodeName = "worker1"
	} else if node_index == 2 {
		nodeName = "worker2"
	}
	return nodeName
}

func (sr Sample) Bind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	fmt.Println("进入Bind........")
	fmt.Printf("pod:%s,namespace:%s\n", pod.Name, pod.Namespace)
	fmt.Printf("nodeName:%v\n", nodeName)
	fmt.Println("-----------------------------------------")
	if pod == nil {
		return framework.NewStatus(framework.Error, fmt.Sprintf("pod cannot be nil"))
	}
	// pod调度到得分最高的节点上
	if nodeName == nodename {
		binding := &v1.Binding{
			ObjectMeta: v12.ObjectMeta{Namespace: pod.Namespace, Name: pod.Name, UID: pod.UID},
			Target:     v1.ObjectReference{Kind: "Node", Name: nodeName},
		}
		t += 1
		err := sr.handle.ClientSet().CoreV1().Pods(pod.Namespace).Bind(ctx, binding, v12.CreateOptions{})
		if err != nil {
			return framework.NewStatus(framework.Error, err.Error())
		}
		return nil
	}
	return nil
}

// type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
func New(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
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
