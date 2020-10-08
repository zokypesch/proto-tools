package commands

import (

	// "gopkg.in/yaml.v2"

	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	core "github.com/zokypesch/proto-lib/core"
	"github.com/zokypesch/proto-tools/config"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Deploy struct for creating proto from DB
type Deploy struct {
	db *gorm.DB
}

var deploy *Deploy

// NewDeploy for new protofrom db
func NewDeploy() CommandInterfacing {

	if deploy == nil {
		cfg := config.Get()
		db := core.InitDB(cfg.DBAddress, cfg.DBName, cfg.DBUser, cfg.DBPassword, cfg.DBPort, false, 10, 5)

		deploy = &Deploy{db}
	}

	return deploy
}

const (
	kubeconfig = `./kubeconfig`
)

// Execute for executing command
func (depl *Deploy) Execute(args map[string]string) error {
	log.Println("Kube tools")

	var ok bool
	var mode string
	var err error

	mode, ok = args["mode"]
	if !ok {
		return fmt.Errorf("Not enough parameter, cmd is required")
	}

	svc, ok := args["svc"]
	if !ok {
		return fmt.Errorf("Not enough parameter, cmd is required")
	}

	env, ok := args["env"] // production or staging
	if !ok {
		return fmt.Errorf("Not enough parameter, env is required")
	}

	version, ok := args["version"]
	if !ok {
		return fmt.Errorf("Not enough parameter, digest is required")
	}

	switch mode {
	case "publish_svc":
		err = depl.pubSvc(svc, env, version, args)
		return err
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	meta, err := GetMeta(depl.db, svc, env, version)
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(config)

	switch mode {
	case "deployment":
		err = depl.deployment(meta, clientset)
	case "service":
		err = depl.service(meta, clientset)
	case "gateway":
		err = depl.gateway(meta, dynClient)
	case "virtual":
		err = depl.virtual(meta, dynClient, args)
	case "scale":
		err = depl.scale(meta, dynClient, args)
	case "push_env":
		err = depl.pushEnv(meta, args)
	case "del_env":
		err = depl.delEnv(meta, args)
	case "publish_digest":
		err = depl.pubDigest(meta, args)
	case "init":
		err = depl.deployment(meta, clientset)
		if err != nil {
			return err
		}
		err = depl.service(meta, clientset)
		if err != nil {
			return err
		}
		err = depl.gateway(meta, dynClient)
		if err != nil {
			return err
		}
		err = depl.virtual(meta, dynClient, args)
	}

	if err != nil {
		return err
	}

	return nil
}

func (depl *Deploy) deployment(meta *KubeSvc, clientset *kubernetes.Clientset) error {
	var err error

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	limitCPU, _ := resourcev1.ParseQuantity("500m")
	limitRAM, _ := resourcev1.ParseQuantity("1Gi")
	requestCPU, _ := resourcev1.ParseQuantity("250m")
	requestRAM, _ := resourcev1.ParseQuantity("512Mi")

	var envarKube []apiv1.EnvVar
	// build env
	for _, envMap := range meta.KubeEnv {
		envarKube = append(envarKube, apiv1.EnvVar{Name: envMap.Name, Value: envMap.Val})
	}

	finalName := meta.Name

	if meta.Version != "v1" {
		finalName = fmt.Sprintf("%s-%s", meta.Name, meta.Version)
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      finalName,
			Namespace: apiv1.NamespaceDefault,
			Labels: map[string]string{
				"environment": meta.Environment,
				"tier":        "backend",
				"version":     meta.Version,
			},
		},
		Spec: appsv1.DeploymentSpec{
			ProgressDeadlineSeconds: int32Ptr(2000),
			Replicas:                int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": meta.Name,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":         meta.Name,
						"environment": meta.Environment,
						"tier":        "backend",
						"version":     meta.Version,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:            meta.Name,
							Image:           fmt.Sprintf("registry-intl.ap-southeast-5.aliyuncs.com/prakerja/%s:%s", meta.Name, meta.KubeHistory.Digest),
							ImagePullPolicy: apiv1.PullAlways,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
								{
									Name:          "grpc",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
							Resources: apiv1.ResourceRequirements{
								Limits: apiv1.ResourceList{
									"cpu":    limitCPU,
									"memory": limitRAM,
								},
								Requests: apiv1.ResourceList{
									"cpu":    requestCPU,
									"memory": requestRAM,
								},
							},
							Env: envarKube,
						},
					},
					DNSPolicy:                     apiv1.DNSClusterFirst,
					RestartPolicy:                 apiv1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: int64Ptr(60),
				},
			},
		},
	}

	_, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {

		newCommand := fmt.Sprintf(`deployments.apps "%s" already exists`, finalName)
		if err.Error() == newCommand {

			// get existing replicate
			resData, err := deploymentsClient.Get(context.TODO(), finalName, metav1.GetOptions{})

			if err != nil {
				return err
			}

			newReplica := resData.Status.Replicas
			deployment.Spec.Replicas = &newReplica

			_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func (depl *Deploy) service(meta *KubeSvc, clientset *kubernetes.Clientset) error {
	svcClient := clientset.CoreV1().Services(apiv1.NamespaceDefault)

	service := &apiv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: meta.Name,
			Labels: map[string]string{
				"app": meta.Name,
			},
			Namespace: apiv1.NamespaceDefault,
		},
		Spec: apiv1.ServiceSpec{
			ExternalTrafficPolicy: apiv1.ServiceExternalTrafficPolicyTypeCluster,
			Ports: []apiv1.ServicePort{
				{
					Name: "http",
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: 80,
						StrVal: "",
					},
					Protocol: apiv1.ProtocolTCP,
				},
				{
					Name: "grpc",
					Port: 8080,
					TargetPort: intstr.IntOrString{
						Type:   0,
						IntVal: 8080,
						StrVal: "",
					},
					Protocol: apiv1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": meta.Name,
			},
			Type: "NodePort",
		},
	}

	_, err := svcClient.Create(context.TODO(), service, metav1.CreateOptions{})

	if err != nil {
		newCommand := fmt.Sprintf(`services "%s" already exists`, meta.Name)
		if err.Error() == newCommand {
			// get resource and cluster
			res, err := svcClient.Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			service.ObjectMeta.ResourceVersion = res.ObjectMeta.ResourceVersion
			service.Spec.ClusterIP = res.Spec.ClusterIP
			// svcClient.Delete(context.TODO(), meta.Name, metav1.DeleteOptions{})
			_, err = svcClient.Update(context.TODO(), service, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			return nil
		}

		return err
	}

	return nil
}

func (depl *Deploy) gateway(meta *KubeSvc, clientset dynamic.Interface) error {
	gatewayRes := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
		Resource: "gateways",
	}

	gw := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-istio-internal-gw", meta.Name),
				"namespace": apiv1.NamespaceDefault,
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"istio": "istio-internal-gw",
				},
				"servers": []map[string]interface{}{
					{
						"port": map[string]interface{}{
							"number":   80,
							"name":     "http",
							"protocol": "HTTP",
						},
						"hosts": []string{
							fmt.Sprintf("%s.prakerja.local", meta.Name),
						},
					},
				},
			},
		},
	}

	_, err := clientset.Resource(gatewayRes).Namespace("default").Create(context.TODO(), gw, metav1.CreateOptions{})
	if err != nil {

		if err.Error() == fmt.Sprintf(`gateways.networking.istio.io "%s-istio-internal-gw" already exists`, meta.Name) {
			res, err := clientset.Resource(gatewayRes).Namespace("default").Get(context.TODO(), fmt.Sprintf("%s-istio-internal-gw", meta.Name), metav1.GetOptions{})
			if err != nil {
				return err
			}
			gw.Object["metadata"].(map[string]interface{})["resourceVersion"] = res.Object["metadata"].(map[string]interface{})["resourceVersion"]
			_, err = clientset.Resource(gatewayRes).Namespace("default").Update(context.TODO(), gw, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			return nil
		}
		return err
	}

	return nil
}

func (depl *Deploy) virtual(meta *KubeSvc, clientset dynamic.Interface, args map[string]string) error {
	gatewayRes := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
		Resource: "virtualservices",
	}

	firstRoute := map[string]interface{}{
		"destination": map[string]interface{}{
			// "host": fmt.Sprintf("%s.default.svc.cluster.local", meta.Name),
			"host": meta.Name,
			"port": map[string]interface{}{
				"number": 80,
			},
			"subset": meta.Version,
		},
		"weight": 100,
	}

	gw := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "VirtualService",
			"metadata": map[string]interface{}{
				"name":      meta.Name,
				"namespace": apiv1.NamespaceDefault,
			},
			"spec": map[string]interface{}{
				"gateways": []string{
					fmt.Sprintf("%s-istio-internal-gw", meta.Name),
				},
				"hosts": []string{
					fmt.Sprintf("%s.prakerja.local", meta.Name),
				},

				"http": []map[string]interface{}{
					{
						"corsPolicy": map[string]interface{}{
							"allowCredentials": true,
							"allowHeaders": []string{
								"DNT",
								"User-Agent",
								"X-Requested-With",
								"If-Modified-Since",
								"Cache-Control",
								"Content-Type",
								"Range",
								"Authorization",
								"authorization",
							},
							"allowMethods": []string{
								"POST",
								"GET",
								"OPTIONS",
								"PUT",
								"DELETE",
							},
							"allowOrigins": []map[string]interface{}{
								{"exact": "*"},
							},
							"maxAge": "24h",
						},
						"match": []map[string]interface{}{
							{
								"uri": map[string]interface{}{
									"prefix": "/",
								},
							},
						},
						"rewrite": map[string]interface{}{
							"uri": "/",
						},
						"route": []map[string]interface{}{
							firstRoute,
						},
					},
				},
			},
		},
	}

	weight, okWeight := args["weight"]

	newWeight := strings.Split(weight, "#")
	mapWeight := make(map[string]int)

	if len(weight) == 0 {
		okWeight = false
	}

	if okWeight {
		for _, vWeightMap := range newWeight {
			splitMap := strings.Split(vWeightMap, ":")

			if len(splitMap) < 2 {
				continue
			}
			newWeightSb, err := strconv.Atoi(splitMap[1])
			if err != nil {
				newWeightSb = 0
			}

			mapWeight[splitMap[0]] = newWeightSb
		}
	}

	_, err := clientset.Resource(gatewayRes).Namespace("default").Create(context.TODO(), gw, metav1.CreateOptions{})
	if err != nil {
		errTxt := fmt.Sprintf(`virtualservices.networking.istio.io "%s" already exists`, meta.Name)
		if err.Error() == errTxt {
			res, err := clientset.Resource(gatewayRes).Namespace("default").Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			gw.Object["metadata"].(map[string]interface{})["resourceVersion"] = res.Object["metadata"].(map[string]interface{})["resourceVersion"]
			lastWeight := 100

			if okWeight {

				var newRoute []map[string]interface{}

				http := res.Object["spec"].(map[string]interface{})["http"].([]interface{})
				routeParse := http[0].(map[string]interface{})
				route := routeParse["route"].([]interface{})

				findSame := false
				for _, vRoute := range route {
					newRouteSingle := vRoute.(map[string]interface{})
					destination := newRouteSingle["destination"].(map[string]interface{})

					resSub, ok := destination["subset"]

					if !ok {
						continue
					}

					if resSub == meta.Version {
						findSame = true
						continue
					}
				}

				if !findSame {
					route = append(route, firstRoute)
				}

				for keyRoute, vRoute := range route {
					newRouteSingle := vRoute.(map[string]interface{})

					destination := newRouteSingle["destination"].(map[string]interface{})
					subSetFound := false

					finalWeight := 0

					if _, ok := destination["subset"]; ok {
						if newMapWeight, ok := mapWeight[destination["subset"].(string)]; ok {
							finalWeight = newMapWeight
							subSetFound = true
						}
					}

					if subSetFound && (lastWeight-finalWeight) >= 0 {
						newRouteSingle["weight"] = finalWeight
					}

					if (keyRoute+1) == len(route) && lastWeight > 0 {
						finalWeight = lastWeight
						newRouteSingle["weight"] = finalWeight
					}

					newRoute = append(newRoute, newRouteSingle)
					lastWeight = lastWeight - finalWeight

				}
				gw.Object["spec"].(map[string]interface{})["http"].([]map[string]interface{})[0]["route"] = newRoute
			}

			// clientset.Resource(gatewayRes).Namespace("default").Delete(context.TODO(), meta.Name, metav1.DeleteOptions{})
			_, err = clientset.Resource(gatewayRes).Namespace("default").Update(context.TODO(), gw, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			return depl.destination(meta, clientset)
		}
		return err
	}

	return depl.destination(meta, clientset)
}

func (depl *Deploy) destination(meta *KubeSvc, clientset dynamic.Interface) error {
	gatewayRule := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
		Resource: "destinationrules",
	}

	res, err := clientset.Resource(gatewayRule).Namespace("default").Get(context.TODO(), meta.Name, metav1.GetOptions{})

	var destination []map[string]interface{}
	foundVersion := false

	if err == nil {
		resDestination := res.Object["spec"].(map[string]interface{})["subsets"].([]interface{})

		for _, v := range resDestination {
			vFinalDest := v.(map[string]interface{})
			if vFinalDest["name"] == meta.Version {
				foundVersion = true
			}
			destination = append(destination, vFinalDest)
		}
	}

	if !foundVersion {
		newDestination := map[string]interface{}{
			"name": meta.Version,
			"labels": map[string]interface{}{
				"version": meta.Version,
			},
		}
		destination = append(destination, newDestination)
	}

	gwRule := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.istio.io/v1beta1",
			"kind":       "DestinationRule",
			"metadata": map[string]interface{}{
				"name":      meta.Name,
				"namespace": apiv1.NamespaceDefault,
			},
			"spec": map[string]interface{}{
				"host":    fmt.Sprintf("%s.default.svc.cluster.local", meta.Name),
				"subsets": destination,
			},
		},
	}
	_, err = clientset.Resource(gatewayRule).Namespace("default").Create(context.TODO(), gwRule, metav1.CreateOptions{})
	if err != nil {
		errTxt := fmt.Sprintf(`destinationrules.networking.istio.io "%s" already exists`, meta.Name)
		if err.Error() == errTxt {

			res, err := clientset.Resource(gatewayRule).Namespace("default").Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			gwRule.Object["metadata"].(map[string]interface{})["resourceVersion"] = res.Object["metadata"].(map[string]interface{})["resourceVersion"]

			_, err = clientset.Resource(gatewayRule).Namespace("default").Update(context.TODO(), gwRule, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			return nil
		}
		return err
	}

	return nil
}

func (depl *Deploy) scale(meta *KubeSvc, clientset dynamic.Interface, args map[string]string) error {

	min, ok := args["min"]
	override := false

	if ok {
		newMin, err := strconv.Atoi(min)

		if err != nil {
			return err
		}

		override = true
		meta.MinScale = newMin
	}

	max, ok := args["max"]

	if ok {
		newMax, err := strconv.Atoi(max)

		if err != nil {
			return err
		}
		override = true
		meta.MaxScale = newMax
	}

	gatewayRes := schema.GroupVersionResource{
		Group:    "autoscaling",
		Version:  "v1",
		Resource: "horizontalpodautoscalers",
	}

	hpa := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "autoscaling/v1",
			"kind":       "HorizontalPodAutoscaler",
			"metadata": map[string]interface{}{
				"name":      meta.Name,
				"namespace": apiv1.NamespaceDefault,
			},
			"spec": map[string]interface{}{
				"maxReplicas": meta.MaxScale,
				"minReplicas": meta.MinScale,
				"scaleTargetRef": map[string]interface{}{
					"apiVersion": "extensions/v1beta1",
					"kind":       "Deployment",
					"name":       meta.Name,
				},
				"targetCPUUtilizationPercentage": 70,
			},
		},
	}

	_, err := clientset.Resource(gatewayRes).Namespace("default").Create(context.TODO(), hpa, metav1.CreateOptions{})
	if err != nil {

		errTxt := fmt.Sprintf(`horizontalpodautoscalers.autoscaling "%s" already exists`, meta.Name)
		if err.Error() == errTxt {
			res, err := clientset.Resource(gatewayRes).Namespace("default").Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			hpa.Object["metadata"].(map[string]interface{})["resourceVersion"] = res.Object["metadata"].(map[string]interface{})["resourceVersion"]
			_, err = clientset.Resource(gatewayRes).Namespace("default").Update(context.TODO(), hpa, metav1.UpdateOptions{})

			if err != nil {
				return err
			}

			if override {
				// update in database
				err = depl.db.Exec("UPDATE kube_svc SET min_scale = ?, max_scale = ? WHERE name = ? AND version = ? AND environment = ? ", min, max, meta.Name, meta.Version, meta.Environment).Error

				if err != nil {
					log.Println("failed to update min, max scale in database")
					return err
				}
			}

			return nil
		}

		return err
	}

	if override {
		// update in database
		err = depl.db.Exec("UPDATE kube_svc SET min_scale = ?, max_scale = ? WHERE name = ? AND version = ? AND environment = ? ", min, max, meta.Name, meta.Version, meta.Environment).Error

		if err != nil {
			log.Println("failed to update min, max scale in database")
			return err
		}
	}

	return nil
}

func (depl *Deploy) pushEnv(meta *KubeSvc, args map[string]string) error {
	name, ok := args["name"]
	if !ok {
		return fmt.Errorf("name cannot be empty")
	}

	value, ok := args["value"]
	if !ok {
		return fmt.Errorf("value cannot be empty")
	}

	var exist []int64
	err := depl.db.Raw("SELECT COUNT(1) AS total FROM kube_env WHERE name = ? AND code = ?", name, meta.Code).Pluck("total", &exist).Error

	if err != nil {
		return err
	}

	if exist[0] > 0 {
		err = depl.db.Exec("UPDATE kube_env SET value = ? WHERE code = ? AND name = ?", value, meta.Code, name).Error

		if err != nil {
			return err
		}

		return nil
	}

	err = depl.db.Exec("INSERT INTO kube_env(code, name, val, version) VALUES(?, ?, ?, 'v1')", meta.Code, name, value).Error
	if err != nil {
		return err
	}
	return nil
}

func (depl *Deploy) pubSvc(svc, env, version string, args map[string]string) error {
	host, ok := args["host"]
	if !ok {
		return fmt.Errorf("host cannot be empty")
	}

	min, ok := args["min"]
	if !ok {
		return fmt.Errorf("min cannot be empty")
	}
	minInt, err := strconv.Atoi(min)
	if err != nil {
		return err
	}

	max, ok := args["max"]
	if !ok {
		return fmt.Errorf("max cannot be empty")
	}
	maxInt, err := strconv.Atoi(max)
	if err != nil {
		return err
	}

	code := fmt.Sprintf("%s_%s_%s", svc, env, version)

	var exist []int64
	err = depl.db.Raw("SELECT COUNT(1) AS total FROM kube_svc WHERE code = ?", code).Pluck("total", &exist).Error

	if err != nil {
		return err
	}

	if exist[0] > 0 {
		err = depl.db.Exec("UPDATE kube_svc SET min_scale = ?, max_scale = ?, host = ? WHERE code = ? ", minInt, maxInt, host, code).Error

		if err != nil {
			return err
		}

		return nil
	}

	err = depl.db.Exec("INSERT INTO kube_svc(code, name, version, environment, host, min_scale, max_scale) VALUES(?, ?, ?, ?, ?, ?, ?)",
		code, svc, version, env, host, minInt, maxInt).Error

	if err != nil {
		return err
	}
	return nil
}

func (depl *Deploy) delEnv(meta *KubeSvc, args map[string]string) error {
	name, ok := args["name"]
	if !ok {
		return fmt.Errorf("name cannot be empty")
	}

	err := depl.db.Exec("DELETE FROM kube_env WHERE code = ? AND name = ?", meta.Code, name).Error

	if err != nil {
		return err
	}
	return nil
}

func (depl *Deploy) pubDigest(meta *KubeSvc, args map[string]string) error {
	digest, ok := args["digest"]
	if !ok {
		return fmt.Errorf("digest cannot be empty")
	}

	err := depl.db.Exec("INSERT INTO kube_history(code, digest) VALUES(?, ?)", meta.Code, digest).Error

	if err != nil {
		return err
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func int64Ptr(i int64) *int64 { return &i }
