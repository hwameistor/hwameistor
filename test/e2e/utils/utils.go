package utils

import (
	"bytes"
	"context"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	b1 "k8s.io/api/batch/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/hwameistor/hwameistor/test/e2e/framework"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func Int32Ptr(i int32) *int32 { return &i }

func BoolPter(i bool) *bool { return &i }

func RunInLinux(cmd string) string {
	result, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logrus.Printf("ERROR:%+v ", err)
	}
	return string(result)
}

func nodeList() *apiv1.NodeList {
	logrus.Printf("get node list")
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	nodelist := &apiv1.NodeList{}
	err := client.List(context.TODO(), nodelist)
	if err != nil {
		logrus.Printf("%+v ", err)
		f.ExpectNoError(err)
	}
	return nodelist
}

func addLabels() {
	logrus.Printf("add node labels")
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	nodelist := &apiv1.NodeList{}
	err := client.List(context.TODO(), nodelist)
	if err != nil {
		f.ExpectNoError(err)
		logrus.Printf("%+v ", err)
	}
	for _, nodes := range nodelist.Items {
		node := &apiv1.Node{}
		nodeKey := k8sclient.ObjectKey{
			Name: nodes.Name,
		}
		err := client.Get(context.TODO(), nodeKey, node)
		if err != nil {
			logrus.Printf("%+v ", err)
			f.ExpectNoError(err)
		}

		if _, exists := node.Labels["lvm.hwameistor.io/enable"]; !exists {
			node.Labels["lvm.hwameistor.io/enable"] = "true"
			logrus.Printf("adding labels ")
			err := client.Update(context.TODO(), node)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
		}

	}
}

func installHwameiStorByHelm() {
	logrus.Infof("helm install hwameistor")
	_ = RunInLinux("helm install hwameistor -n hwameistor ../../helm/hwameistor --create-namespace --set global.k8sImageRegistry=k8s-gcr.m.daocloud.io")
}

func installHwameiStorByHelm_offline() {
	logrus.Infof("helm install hwameistor")
	_ = RunInLinux("helm install hwameistor -n hwameistor ../../helm/hwameistor --create-namespace --set global.k8sImageRegistry=172.30.45.210")
}

func StartAdRollback(k8s string) error {
	if k8s == "kylin10arm" {
		logrus.Info("start arm_rollback")
		run := "sh arm_rollback.sh "
		_ = RunInLinux(run)
	} else {
		logrus.Info("start ad_rollback" + k8s)
		run := "sh ad_rollback.sh " + k8s
		_ = RunInLinux(run)
	}

	err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		output := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
		if output != "1\n" {
			return false, nil
		} else {
			logrus.Info("k8s ready")
			return true, nil
		}

	})
	if err != nil {
		logrus.Error(err)
	}
	return err

}

func ConfigureadEnvironment(ctx context.Context, k8s string) error {

	if k8s == "centos7.9_offline" {
		installHwameiStorByHelm_offline()
	} else {
		installHwameiStorByHelm()
	}

	installDrbd()
	addLabels()
	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()

	drbd1 := &b1.Job{}
	drbdKey1 := k8sclient.ObjectKey{
		Name:      "drbd-adapter-k8s-node1-rhel7",
		Namespace: "hwameistor",
	}
	drbd2 := &b1.Job{}
	drbdKey2 := k8sclient.ObjectKey{
		Name:      "drbd-adapter-k8s-node2-rhel7",
		Namespace: "hwameistor",
	}

	localStorage := &appsv1.DaemonSet{}
	localStorageKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage",
		Namespace: "hwameistor",
	}
	err := client.Get(ctx, localStorageKey, localStorage)

	controller := &appsv1.Deployment{}
	controllerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage-csi-controller",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, controllerKey, controller)

	webhook := &appsv1.Deployment{}
	webhookKey := k8sclient.ObjectKey{
		Name:      "hwameistor-admission-controller",
		Namespace: "hwameistor",
	}

	scheduler := &appsv1.Deployment{}
	schedulerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-scheduler",
		Namespace: "hwameistor",
	}

	localDiskManager := &appsv1.DaemonSet{}
	localDiskManagerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-disk-manager",
		Namespace: "hwameistor",
	}

	logrus.Infof("waiting for drbd ready")

	err = wait.PollImmediate(3*time.Second, 15*time.Minute, func() (done bool, err error) {
		err1 := client.Get(ctx, drbdKey1, drbd1)
		err2 := client.Get(ctx, drbdKey2, drbd2)

		if k8serror.IsNotFound(err1) && k8serror.IsNotFound(err2) {
			return true, nil
		}
		return false, nil
	})

	logrus.Infof("waiting for hwamei ready")

	err = wait.PollImmediate(3*time.Second, 20*time.Minute, func() (done bool, err error) {
		err = client.Get(ctx, localStorageKey, localStorage)
		if err != nil {
			logrus.Error(" localStorage error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, controllerKey, controller)
		if err != nil {
			logrus.Error("controller error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, schedulerKey, scheduler)
		if err != nil {
			logrus.Error("scheduler error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, localDiskManagerKey, localDiskManager)
		if err != nil {
			logrus.Error("localDiskManager error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, webhookKey, webhook)
		if err != nil {
			logrus.Error("admission-controller error ", err)
			f.ExpectNoError(err)
		}

		if localStorage.Status.DesiredNumberScheduled == localStorage.Status.NumberAvailable && controller.Status.AvailableReplicas == int32(1) && scheduler.Status.AvailableReplicas == int32(1) && localDiskManager.Status.DesiredNumberScheduled == localDiskManager.Status.NumberAvailable && webhook.Status.AvailableReplicas == int32(1) {
			return true, nil
		}
		return false, nil
	})

	return err
}

func ConfigureEnvironment(ctx context.Context) error {
	logrus.Info("start rollback")
	_ = RunInLinux("sh rollback.sh")
	err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		output := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
		if output != "1\n" {
			return false, nil
		} else {
			logrus.Info("k8s ready")
			return true, nil
		}

	})
	if err != nil {
		logrus.Error(err)
	}
	installHwameiStorByHelm()
	installDrbd()
	if err != nil {
		logrus.Error(err)
	}
	addLabels()
	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()

	drbd1 := &b1.Job{}
	drbdKey1 := k8sclient.ObjectKey{
		Name:      "drbd-adapter-k8s-node1-rhel7",
		Namespace: "hwameistor",
	}
	drbd2 := &b1.Job{}
	drbdKey2 := k8sclient.ObjectKey{
		Name:      "drbd-adapter-k8s-node2-rhel7",
		Namespace: "hwameistor",
	}

	localStorage := &appsv1.DaemonSet{}
	localStorageKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, localStorageKey, localStorage)

	controller := &appsv1.Deployment{}
	controllerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage-csi-controller",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, controllerKey, controller)

	webhook := &appsv1.Deployment{}
	webhookKey := k8sclient.ObjectKey{
		Name:      "hwameistor-admission-controller",
		Namespace: "hwameistor",
	}

	scheduler := &appsv1.Deployment{}
	schedulerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-scheduler",
		Namespace: "hwameistor",
	}

	localDiskManager := &appsv1.DaemonSet{}
	localDiskManagerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-disk-manager",
		Namespace: "hwameistor",
	}

	logrus.Infof("waiting for drbd ready")

	err = wait.PollImmediate(3*time.Second, 15*time.Minute, func() (done bool, err error) {
		err1 := client.Get(ctx, drbdKey1, drbd1)
		err2 := client.Get(ctx, drbdKey2, drbd2)

		if k8serror.IsNotFound(err1) && k8serror.IsNotFound(err2) {
			return true, nil
		}
		return false, nil
	})

	logrus.Infof("waiting for hwamei ready")

	err = wait.PollImmediate(3*time.Second, 20*time.Minute, func() (done bool, err error) {
		err = client.Get(ctx, localStorageKey, localStorage)
		if err != nil {
			logrus.Error(" localStorage error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, controllerKey, controller)
		if err != nil {
			logrus.Error("controller error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, schedulerKey, scheduler)
		if err != nil {
			logrus.Error("scheduler error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, localDiskManagerKey, localDiskManager)
		if err != nil {
			logrus.Error("localDiskManager error ", err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, webhookKey, webhook)
		if err != nil {
			logrus.Error("admission-controller error ", err)
			f.ExpectNoError(err)
		}

		if localStorage.Status.DesiredNumberScheduled == localStorage.Status.NumberAvailable && controller.Status.AvailableReplicas == int32(1) && scheduler.Status.AvailableReplicas == int32(1) && localDiskManager.Status.DesiredNumberScheduled == localDiskManager.Status.NumberAvailable && webhook.Status.AvailableReplicas == int32(1) {
			return true, nil
		}
		return false, nil
	})

	return err
}

func ConfigureEnvironmentForPrTest(ctx context.Context) bool {
	err := wait.PollImmediate(10*time.Second, 10*time.Minute, func() (done bool, err error) {
		output := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
		if output != "1\n" {
			return false, nil
		} else {
			logrus.Info("k8s ready")
			return true, nil
		}

	})
	if err != nil {
		logrus.Error(err)
	}
	installHwameiStorByHelm()
	addLabels()
	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()

	localStorage := &appsv1.DaemonSet{}
	localStorageKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, localStorageKey, localStorage)
	if err != nil {
		logrus.Error("%+v ", err)
		f.ExpectNoError(err)
	}

	controller := &appsv1.Deployment{}
	controllerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-storage-csi-controller",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, controllerKey, controller)
	if err != nil {
		logrus.Error(err)
		f.ExpectNoError(err)
	}
	webhook := &appsv1.Deployment{}
	webhookKey := k8sclient.ObjectKey{
		Name:      "hwameistor-admission-controller",
		Namespace: "hwameistor",
	}
	err = client.Get(ctx, webhookKey, webhook)
	if err != nil {
		logrus.Error(err)
		f.ExpectNoError(err)
	}

	scheduler := &appsv1.Deployment{}
	schedulerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-scheduler",
		Namespace: "hwameistor",
	}

	err = client.Get(ctx, schedulerKey, scheduler)
	if err != nil {
		logrus.Error(err)
		f.ExpectNoError(err)
	}
	localDiskManager := &appsv1.DaemonSet{}
	localDiskManagerKey := k8sclient.ObjectKey{
		Name:      "hwameistor-local-disk-manager",
		Namespace: "hwameistor",
	}

	err = client.Get(ctx, localDiskManagerKey, localDiskManager)
	if err != nil {
		logrus.Error(err)
		f.ExpectNoError(err)

	}

	logrus.Infof("waiting for ready")
	ch := make(chan struct{}, 1)
	go func() {
		for localStorage.Status.DesiredNumberScheduled != localStorage.Status.NumberAvailable || controller.Status.AvailableReplicas != int32(1) || scheduler.Status.AvailableReplicas != int32(1) || localDiskManager.Status.DesiredNumberScheduled != localDiskManager.Status.NumberAvailable || webhook.Status.AvailableReplicas != int32(1) {
			time.Sleep(10 * time.Second)
			err := client.Get(ctx, localStorageKey, localStorage)
			if err != nil {
				logrus.Error(" localStorage error ", err)
				f.ExpectNoError(err)
			}
			err = client.Get(ctx, controllerKey, controller)
			if err != nil {
				logrus.Error("controller error ", err)
				f.ExpectNoError(err)
			}
			err = client.Get(ctx, schedulerKey, scheduler)
			if err != nil {
				logrus.Error("scheduler error ", err)
				f.ExpectNoError(err)
			}
			err = client.Get(ctx, localDiskManagerKey, localDiskManager)
			if err != nil {
				logrus.Error("localDiskManager error ", err)
				f.ExpectNoError(err)
			}
			err = client.Get(ctx, webhookKey, webhook)
			if err != nil {
				logrus.Error("admission-controller error ", err)
				f.ExpectNoError(err)
			}

		}
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		logrus.Infof("Components are ready ")
		return true
	case <-time.After(20 * time.Minute):
		logrus.Error("timeout")
		return false

	}

}

func UninstallHelm() {
	logrus.Printf("helm uninstall hwameistor")
	_ = RunInLinux("helm list -A | grep 'hwameistor' | awk '{print $1}' | xargs helm uninstall -n hwameistor")
	logrus.Printf("clean all hwameistor crd")
	f := framework.NewDefaultFramework(extv1.AddToScheme)
	client := f.GetClient()
	crdList := extv1.CustomResourceDefinitionList{}
	err := client.List(context.TODO(), &crdList)
	if err != nil {
		logrus.Printf("%+v ", err)
		f.ExpectNoError(err)
	}
	for _, crd := range crdList.Items {
		myBool, _ := regexp.MatchString(".*hwameistor.*", crd.ObjectMeta.Name)
		if myBool {
			err := client.Delete(context.TODO(), &crd)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
		}

	}
	logrus.Printf("waiting for uninstall hwameistor")

}

func CreateLdc(ctx context.Context) error {
	logrus.Printf("create ldc for each node")
	nodelist := nodeList()
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	for _, nodes := range nodelist.Items {
		exmlocalDiskClaim := &v1alpha1.LocalDiskClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "localdiskclaim-" + nodes.Name,
				Namespace: "kube-system",
			},
			Spec: v1alpha1.LocalDiskClaimSpec{
				NodeName: nodes.Name,
				Description: v1alpha1.DiskClaimDescription{
					DiskType: "HDD",
				},
			},
		}
		err := client.Create(ctx, exmlocalDiskClaim)
		if err != nil {
			logrus.Printf("Create LDC failed ：%+v ", err)
			f.ExpectNoError(err)
		}
	}

	err := wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
		for _, nodes := range nodelist.Items {
			time.Sleep(3 * time.Second)
			localDiskClaim := &v1alpha1.LocalDiskClaim{}
			localDiskClaimKey := k8sclient.ObjectKey{
				Name:      "localdiskclaim-" + nodes.Name,
				Namespace: "kube-system",
			}
			err := client.Get(ctx, localDiskClaimKey, localDiskClaim)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			if localDiskClaim.Status.Status != v1alpha1.LocalDiskClaimStatusBound {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		logrus.Error(err)
		return err
	} else {
		return nil
	}

}

func DeleteAllPVC(ctx context.Context) error {
	logrus.Printf("delete All PVC")
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	pvcList := &apiv1.PersistentVolumeClaimList{}
	err := client.List(ctx, pvcList)
	if err != nil {
		logrus.Error("get pvc list error ", err)
		f.ExpectNoError(err)
	}

	for _, pvc := range pvcList.Items {
		logrus.Printf("delete pvc:%+v ", pvc.Name)
		ctx, _ := context.WithTimeout(ctx, time.Minute)
		err := client.Delete(ctx, &pvc)
		if err != nil {
			logrus.Error("delete pvc error: ", err)
			f.ExpectNoError(err)
		}
	}

	err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
		err = client.List(ctx, pvcList)
		if err != nil {
			logrus.Error("get pvc list error: ", err)
			f.ExpectNoError(err)
		}
		if len(pvcList.Items) != 0 {
			return false, nil
		} else {
			return true, nil
		}
	})
	if err != nil {
		logrus.Error(err)
		return err
	} else {
		return nil
	}

}

func DeleteAllSC(ctx context.Context) error {
	logrus.Printf("delete All SC")
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	scList := &storagev1.StorageClassList{}
	err := client.List(ctx, scList)
	if err != nil {
		logrus.Error("get sc list error:", err)
		f.ExpectNoError(err)
	}

	for _, sc := range scList.Items {
		logrus.Printf("delete sc:%+v ", sc.Name)
		err := client.Delete(ctx, &sc)
		if err != nil {
			logrus.Error("delete sc error", err)
			f.ExpectNoError(err)
		}
	}
	err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
		err = client.List(ctx, scList)
		if err != nil {
			logrus.Error("get sc list error", err)
			f.ExpectNoError(err)
		}
		if len(scList.Items) != 0 {
			return false, nil
		} else {
			return true, nil
		}
	})
	if err != nil {
		logrus.Error(err)
		return err
	} else {
		return nil
	}

}

func ExecInPod(config *rest.Config, namespace, podName, command, containerName string) (string, string, error) {
	k8sCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", "", err
	}
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	const tty = false
	req := k8sCli.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).SubResource("exec").Param("container", containerName)
	req.VersionedParams(
		&v1.PodExecOptions{
			Command: cmd,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     tty,
		},
		scheme.ParameterCodec,
	)

	var stdout, stderr bytes.Buffer
	myExec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	err = myExec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}

func installDrbd() {
	logrus.Printf("installing drbd")
	_ = RunInLinux("sh install_drbd.sh")

}
