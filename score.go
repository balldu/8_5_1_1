package plugins

import (
	"context"
	"fmt"
	"math"
	"math/rand"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	//加入随机森林的依赖
	"github.com/yugecode/custom-scheduler/pkg/forest"
)

var t_a int

const S_Name = "custom-plugin"

func (s *Sample) S_Name() string {
	return S_Name
}

func (a *Args_ATH) ATH() [][3]int {
	//// 服务器的资源余量，计算y_i
	//var y_i float64 = 1.0 / 3
	// 存储互补应用集合P_a
	var P_a [][3]int = make([][3]int, a.N)
	for a1 := 0; a1 < a.N; a1++ {
		// 存储每个应用的VD值
		var VD []float64 = make([]float64, a.N)
		// temp存放应用的总开销
		temp := a.a_r[a1][0] + a.a_r[a1][1] + a.a_r[a1][2]
		// 相似度最低，VD值最高的三个应用
		var apps [3]int = [3]int{0, 0, 0}
		// 从大到小排列
		var apps_vd [3]float64 = [3]float64{0, 0, 0}
		for a2 := a1 + 1; a2 < a.N; a2++ {
			temp2 := a.a_r[a2][0] + a.a_r[a2][1] + a.a_r[a2][2]
			VD[a2] += math.Pow(a.a_r[a1][0]/temp-a.a_r[a2][0]/temp2, 2) + math.Pow(a.a_r[a1][1]/temp-a.a_r[a2][1]/temp2, 2) +
				math.Pow(a.a_r[a1][2]/temp-a.a_r[a2][2]/temp2, 2)
			if VD[a2] >= apps_vd[0] {
				apps_vd[2] = apps_vd[1]
				apps_vd[1] = apps_vd[0]
				apps_vd[0] = VD[a2]
				apps[2] = apps[1]
				apps[1] = apps[0]
				apps[0] = a2
			} else if VD[a2] >= apps_vd[1] {
				apps_vd[2] = apps_vd[1]
				apps_vd[1] = VD[a2]
				apps[2] = apps[1]
				apps[1] = a2
			} else if VD[a2] >= apps_vd[2] {
				apps_vd[2] = VD[2]
				apps[2] = a2
			}
		}
		P_a[a1][0] = apps[0]
		P_a[a1][1] = apps[1]
		P_a[a1][2] = apps[2]
	}
	return P_a
}

func (a *Args_ATH) Compute_VD(C []float64) []float64 {
	VD := make([]float64, a.M)
	for i := 0; i < a.M; i++ {
		// temp、temp2存放各自的总开销
		temp := C[0] + C[1] + C[2]
		temp2 := a.m_total_resources[i][0] + a.m_total_resources[i][1] + a.m_total_resources[i][2]
		VD[i] = math.Pow(C[0]/temp-a.m_total_resources[i][0]/temp2, 2) +
			math.Pow(C[1]/temp-a.m_total_resources[i][1]/temp2, 2) +
			math.Pow(C[1]/temp-a.m_total_resources[i][2]/temp2, 2)
	}
	return VD
}

// Score 回调方法
func (sr Sample) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	fmt.Println("进入Score........")
	fmt.Printf("pod:%s,namespace:%s\n", pod.Name, pod.Namespace)
	fmt.Printf("nodeName:%v\n", nodeName)
	fmt.Println("-----------------------------------------")
	// 表征租户t，计算C_t, 获取请求的应用t_a,即pod所请求的应用类型
	var tl float64
	t_a = rand.Intn(a.N)
	a.t_a[t_a] = 1
	C_t := make([]float64, 3)
	C_t = []float64{float64(a.tenant[t]) * a.a_r[t_a][0], float64(a.tenant[t]) * a.a_r[t_a][1], float64(a.tenant[t]) * a.a_r[t_a][0]}
	Res_m := a.Compute_VD(C_t)
	min := Res_m[0]
	//fmt.Println(a.A_a_m)
	for i := 0; i < a.M; i++ {
		if Res_m[i] <= min && a.A_a_m[t_a][i] == 1 {
			a.m_star = i
			min = Res_m[i]
		}
	}
	if a.m_star != -1 {
		a.m_total_resources[a.m_star][0] = a.m_total_resources[a.m_star][0] - C_t[0]
		a.m_total_resources[a.m_star][1] = a.m_total_resources[a.m_star][1] - C_t[0]
		a.m_total_resources[a.m_star][2] = a.m_total_resources[a.m_star][2] - C_t[0]
		a.Num += 1
		a.w_m += 1
		a.tenants -= 1
		goto step_3
	} else {
		fmt.Println("需要进去Score_Extended........")
		score, status := sr.Score_Extended()
		return score, status
	}
step_3:
	tl = math.Max(a.m_total_resources[a.m_star][0], math.Max(a.m_total_resources[a.m_star][1], a.m_total_resources[a.m_star][2])) -
		math.Min(a.m_total_resources[a.m_star][0], math.Min(a.m_total_resources[a.m_star][1], a.m_total_resources[a.m_star][2]))
	if a.w_m >= a.M && tl >= a.gamma {
		Res_a := a.Compute_VD(a.C_application)
		max := Res_a[0]
		for i := range Res_a {
			if Res_a[i] >= max && a.A_a_m[i][a.m_star] == 0 {
				a.a_star = i
			}
		}
		if a.a_star != -1 {
			a.A_a_m[a.a_star][a.m_star] = 1
			a.record[a.a_star] = 1
			a.w_m = 0
		}
	}
	if a.m_star == 0 {
		// 100表示直接就能够找到合适的服务器节点
		fmt.Println("得出结果........k8s-master")
		score := sr.getScore(state, "master")
		return score, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-master"))
	} else if a.m_star == 1 {
		fmt.Println("得出结果........k8s-worker1")
		score := sr.getScore(state, "worker1")
		return score, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-worker1"))
	} else if a.m_star == 2 {
		fmt.Println("得出结果........k8s-worker2")
		score := sr.getScore(state, "worker2")
		return score, framework.NewStatus(framework.Success, fmt.Sprintf("节点 %s 成功符合标准", "k8s-worker2"))
	}
	fmt.Println("得出结果........None")
	return 1, framework.NewStatus(framework.Unschedulable, fmt.Sprintf("目前没有找到合适的节点"))
}

// 注意更新 ScoreExtensions 方法来返回包含 NormalizeScore 方法实现的实例。
func (sr *Sample) ScoreExtensions() framework.ScoreExtensions {
	fmt.Println("进入ScoreExtensions........")
	fmt.Println("-----------------------------------------")
	return sr
}

// type PluginFactory = func(configuration *runtime.Unknown, f FrameworkHandle) (Plugin, error)
func New_S(configuration *runtime.Unknown, f framework.FrameworkHandle) (framework.Plugin, error) {
	args := &Args{}
	if err := framework.DecodeInto(configuration, args); err != nil {
		return nil, err
	}
	klog.V(3).Infof("get plugin config args: %+v", args)
	// 创建随机森林模型
	reg1 := forest.NewRegressor(
		forest.NumTrees(10),
		forest.MaxFeatures(3),
		forest.MinSplit(2),
		forest.MinLeaf(1),
		forest.MaxDepth(10),
		forest.NumWorkers(1),
	)
	// 创建随机森林模型
	reg2 := forest.NewRegressor(
		forest.NumTrees(10),
		forest.MaxFeatures(3),
		forest.MinSplit(2),
		forest.MinLeaf(1),
		forest.MaxDepth(10),
		forest.NumWorkers(1),
	)
	// 创建随机森林模型
	reg3 := forest.NewRegressor(
		forest.NumTrees(10),
		forest.MaxFeatures(3),
		forest.MinSplit(2),
		forest.MinLeaf(1),
		forest.MaxDepth(10),
		forest.NumWorkers(1),
	)
	// 使用训练数据进行训练
	reg1.Fit(featureData, Num_GPU)
	reg2.Fit(featureData, GMem)
	reg3.Fit(featureData, Bandwidth)

	return &Sample{
		args:   args,
		handle: f,
		model1: reg1,
		model2: reg2,
		model3: reg3,
	}, nil
}
