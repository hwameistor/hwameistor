package utils

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"reflect"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	b1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
)

func Int32Ptr(i int32) *int32 { return &i }

func BoolPter(i bool) *bool { return &i }

func RunInLinux(cmd string) (string, error) {
	result, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		logrus.Printf("ERROR:%+v ", err)
	}
	return string(result), err
}

func nodeList() *corev1.NodeList {
	logrus.Printf("get node list")
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	nodelist := &corev1.NodeList{}
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
	nodelist := &corev1.NodeList{}
	err := client.List(context.TODO(), nodelist)
	if err != nil {
		f.ExpectNoError(err)
		logrus.Printf("%+v ", err)
	}
	for _, nodes := range nodelist.Items {
		node := &corev1.Node{}
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

func installHwameiStorByHelm() error {
	logrus.Infof("helm install hwameistor")
	_, err := RunInLinux("helm install hwameistor -n hwameistor ../../helm/hwameistor --create-namespace --set global.k8sImageRegistry=m.daocloud.io/registry.k8s.io")
	return err

}

func installHwameiStorByHelm_offline() error {
	logrus.Infof("helm install hwameistor")
	_, err := RunInLinux("helm install hwameistor -n hwameistor ../../helm/hwameistor --create-namespace --set global.k8sImageRegistry=10.6.112.210")
	return err
}

func StartAdRollback(k8s string) error {
	if k8s == "kylin10arm" {
		logrus.Info("start arm_rollback")
		run := "sh arm_rollback.sh "
		_, _ = RunInLinux(run)
	} else {
		logrus.Info("start ad_rollback" + k8s)
		run := "sh ad_rollback.sh " + k8s
		_, _ = RunInLinux(run)
	}

	err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		output, _ := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
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
		err := installHwameiStorByHelm_offline()
		if err != nil {
			logrus.Printf(" installHwameiStorByHelm_offline ERROR:%+v ", err)
			return err
		}
	} else {
		err := installHwameiStorByHelm()
		if err != nil {
			logrus.Printf(" installHwameiStorByHelm ERROR:%+v ", err)
			return err
		}
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
		} else if drbd1.Status.Succeeded == int32(1) && drbd2.Status.Succeeded == int32(1) {
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
	_, _ = RunInLinux("sh rollback.sh")

	err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
		output, _ := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
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

	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()

	err = installHwameiStorByHelm()
	if err != nil {
		logrus.Printf(" installHwameiStorByHelm ERROR:%+v ", err)
		return err
	}
	installDrbd()
	if err != nil {
		logrus.Error(err)
	}

	secret_data := map[string][]byte{
		"tls.crt": String2Bytes("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZzVENDQTVtZ0F3SUJBZ0lDQm5vd0RRWUpLb1pJaHZjTkFRRUxCUUF3R0RFV01CUUdBMVVFQ2hNTmFIZGgKYldWcGMzUnZjaTVwYnpBZUZ3MHlOREEwTVRVd056UTFNRE5hRncweU5UQTBNVFV3TnpRMU1ETmFNRkV4RmpBVQpCZ05WQkFvVERXaDNZVzFsYVhOMGIzSXVhVzh4TnpBMUJnTlZCQU1UTG1oM1lXMWxhWE4wYjNJdFlXUnRhWE56CmFXOXVMV052Ym5SeWIyeHNaWEl1YUhkaGJXVnBjM1J2Y2k1emRtTXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUEKQTRJQ0R3QXdnZ0lLQW9JQ0FRRE9mTmlqWnErWTNlSHF2Z0VpOGpLdGtsSnp2RjM0YzFqUVQ5YnFBM3RnUjIvNApVbWQ0dmJKcFIzUUhZL1RRYjMwZS93SVp5OTdyNVA4SmJkS1RNWmMrdEI0WTJtK1c2YXJhMmd5UDBUQzlHTHpOCis5VzNlTnhheWh3Y1VTK08rQTc2Sll0MkVwcVp3TFd4bjA1bHc0OW16STVkbFNnMXV2WTcvYlhYQjFQVFlDajEKTVk2c2VhNXI3TitHSFluSG1GSEJTWi96UE02TEhsaUQ5ay9hQm4wVXRrR09TYkpxc0xBUFptOWg4SXFveDNYUwpEQTRWOG1rb2xTNDliMW5QL0RLekpGdFNySllaK2lyYXl2NU9aWURvblRVMCs5b09IbFFZNjZvNUVlellzZGtVCkhPUXQzazRLWTJEUERkNy9TVWt0ak9SL2ZYdGVBeEIwTDhuSi9xaHREOXdJQnFGbDVsQ1ZGeW51RlJKTVJMekYKVDVvcVluZGxuTjBLMWxTc3lBUzNkV2R3T25vdzhrVGlKbVRPTE1VdU5pV2tLcktmdnpzMENENGlsTUp6YU1sQwpBU0lSOU43eUZrN0FSTHREWlRSa2M1SmhCOStlR1pnYjdnd29CZ2FRTm43QmlNcmNxZG5JL0FDT3IwSmlCOFVYCkpBQUpWQ1kySDlja0tHUjZ6dGpYUzh0cTlHYnR1OTZzbGdpN05yakUvaElvQmZyZllJMkFWSDExSHZBU2dlMjcKMFkycnFqSXBrTVdxQzBSVUdleVV0alpld2EvUGFWU0Y0N0xpY0E0WlI1TDVadGxiRnNFSStxbEgxb3FONmNKSQpuL2YvclNaeVBpZmJGZmRLL1YxTWdQeHl3ajEyKzkwQlFsdjYyYXdWTmdJNmpPTHpTZUJMQTZzSTRGWnR4d0lECkFRQUJvNEhMTUlISU1BNEdBMVVkRHdFQi93UUVBd0lIZ0RBZEJnTlZIU1VFRmpBVUJnZ3JCZ0VGQlFjREFnWUkKS3dZQkJRVUhBd0V3RGdZRFZSME9CQWNFQlFFQ0F3UUdNSUdHQmdOVkhSRUVmekI5Z2g5b2QyRnRaV2x6ZEc5eQpMV0ZrYldsemMybHZiaTFqYjI1MGNtOXNiR1Z5Z2lwb2QyRnRaV2x6ZEc5eUxXRmtiV2x6YzJsdmJpMWpiMjUwCmNtOXNiR1Z5TG1oM1lXMWxhWE4wYjNLQ0xtaDNZVzFsYVhOMGIzSXRZV1J0YVhOemFXOXVMV052Ym5SeWIyeHMKWlhJdWFIZGhiV1ZwYzNSdmNpNXpkbU13RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUJjMTdWd0NwY3RBYXBLOAptVHNuWlNtOWw5OC93TUp6eGhGRlpYUTNVVjBaRkozZ01SYThwZ2NZcTZ6bGl4YWNjdWJQeHhEWmtWNUwwVmJpCllwN1VuTEVITlRyLytaQTlnVDZGWDU2ekIxTFhMcFI5SVRwVzB1Z1ZKeUtrK3ExOWJrektkdzdGN0szT2lvU3MKT09YNm5WZHFqVHhkTkE2TVA5b2ZXN08xYXBTY1FWMTYwUmoyUXpINW9BZFJncUFkNEduVTkzaGQ5VmtJU0FRZwpUQ21MSWZwYjJHbVQrTzUxTVJsckVRSzl6dVJKSXRQS28zalNkUi9pUlVBdVBzZTNmaldHZmZtYU43b3B2bWVwCllqZWpjMFpTZHpEQTlkTDRwSW9PSWdVQUxDb1Y5aU4xSkhLSDl4NnJkMnFQdjRURUFEN25LcXNxNmVKbXVrQ2QKeEJjQUN4SkpYTkplODUwSDd5cDRxWWQxbjRNSEE1emtCWlZPN2htWHUxT2pOMFJrWFdyUGtNZDh4NDZjbnU4bApnWENNWXU2VEUwZFpEVVpFbjBwV25qL29tYWNuZ2VUZGpLcjAxMmVDR0ZLSjdYT1ladWRCcEdKcWtHNHZSVlBHCnAwdGJoU3dVeFJNQzlQcDc5cGJncnhjbXJZdHpqcjR0RU12Qk56azI3QUlNQVRSb0Y5YmxLVzdLNWtsdmNoc3UKMitSOXhsbkVPcHlXdUsrUTJ6ZVc0Vm5uY3ZDaEFYdU1Kc1hSeHNlOXlUbjh4OWdieEJqRlJTcFZXV2dUZGl1cQpSUlQ4NHNDbmRkbjRWSHl4eUMraHpOUk1FalNGYUZpSm16UnZHVzdjbmdEZkcrbGVwRlNaVFpyS2NYa3B4TXBvCkdlbzU2b3o1d2lUTzZ3c0lwblBxTHZaeU5ISTkKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="),
		"tls.key": String2Bytes("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBem56WW8yYXZtTjNoNnI0Qkl2SXlyWkpTYzd4ZCtITlkwRS9XNmdON1lFZHYrRkpuCmVMMnlhVWQwQjJQMDBHOTlIdjhDR2N2ZTYrVC9DVzNTa3pHWFByUWVHTnB2bHVtcTJ0b01qOUV3dlJpOHpmdlYKdDNqY1dzb2NIRkV2anZnTytpV0xkaEthbWNDMXNaOU9aY09QWnN5T1haVW9OYnIyTy8yMTF3ZFQwMkFvOVRHTwpySG11YSt6ZmhoMkp4NWhSd1VtZjh6ek9peDVZZy9aUDJnWjlGTFpCamtteWFyQ3dEMlp2WWZDS3FNZDEwZ3dPCkZmSnBLSlV1UFc5Wnovd3lzeVJiVXF5V0dmb3Eyc3IrVG1XQTZKMDFOUHZhRGg1VUdPdXFPUkhzMkxIWkZCemsKTGQ1T0NtTmd6dzNlLzBsSkxZemtmMzE3WGdNUWRDL0p5ZjZvYlEvY0NBYWhaZVpRbFJjcDdoVVNURVM4eFUrYQpLbUozWlp6ZEN0WlVyTWdFdDNWbmNEcDZNUEpFNGlaa3ppekZMallscENxeW43ODdOQWcrSXBUQ2MyakpRZ0VpCkVmVGU4aFpPd0VTN1EyVTBaSE9TWVFmZm5obVlHKzRNS0FZR2tEWit3WWpLM0tuWnlQd0FqcTlDWWdmRkZ5UUEKQ1ZRbU5oL1hKQ2hrZXM3WTEwdkxhdlJtN2J2ZXJKWUl1emE0eFA0U0tBWDYzMkNOZ0ZSOWRSN3dFb0h0dTlHTgpxNm95S1pERnFndEVWQm5zbExZMlhzR3Z6MmxVaGVPeTRuQU9HVWVTK1diWld4YkJDUHFwUjlhS2plbkNTSi8zCi82MG1jajRuMnhYM1N2MWRUSUQ4Y3NJOWR2dmRBVUpiK3Rtc0ZUWUNPb3ppODBuZ1N3T3JDT0JXYmNjQ0F3RUEKQVFLQ0FnQlRBdmZwQVFBNUdZS1laNWgzWHgwbWtIdWJVYkFKa1AyMGlWUmlNZElveTY5bENtUFVzNWFGSXBldgozT3h2MjFKVzRyZ0tmTjU2d1U0U2tWY2xtN3N0RnlZK1lacVk5Z21kQUNvcXF6UEg0YWtOdWwrWjZmblVlZ3BCCkRCWnRpS1lreEhGRG40a05wTjZtc04vSW80MEVDOXAwSHVIU2IvNUw4ZHhudk9kalEvWlRWSUFDTU0wQlVISXcKTGJIVlR4UVcrNjFWRVM0TDJjQkdGUURYS3VLZ01BRTgzYjNmakhiTFNpZ2ZISi94NkhZaVlueExNSEE4UTNVYQpkNmlJcCtBT0N1QzM1UEJYTXk2RlZUUlZlaE1QOFJ1UjJHbzZZTkQ0VDgzWU1hK2hFaWF1SWt0TmNXQzRabWxsCjlES2hPT28xYUwzM3p3K0lwUHk1YVBqYnZ3Z2dUbXRHWDhldTJ5K3NiMWxOSUZzK0RzYzY0WDVuWFMxR2drTzUKRWEwNngvM0E1ajRVYWVKRXdvSnVRNkNBdEwvUGUrN2N0d1hBdExRV2s3M1k2YStsOEhrbERBTmV1SDFkWVZvTwpVQ2NEUGxuZFk3dEt0NUtSeXdEcUU5N0gzeEhTU3g1R1AwSUJqSEN6NVNDNlg2Vk94YlJSbEdjQ2NsOS8raGdDCllSNWFGcFVqV0FrWTlzNEE3bEdhVXl0M3VrQVFMSCt4OUl3R0wyQUx6Ujdzb1QxMUpZZEhyWVBwSGZJUFA4MXkKVUJ1VytubTk3QzNlamV5d1RPNTMvRkJMeTRqdkRqWk9ETXEySFJmbmhjMGMxbmhlcTlEN1lFNnBEL3Nna0wwagpoazgwaFJGbEw3SGVGQU9hd1lsM2xBY01WTFgzUjZDMFFVLytRTDNzOU1oRzhIZVFxUUtDQVFFQXp1eUJ6VldiCnUrbnJLdUxmcCt0TGhBZWJMSEpuNmJNOFBFb2h3YUMxMUdLajh3eW9nZW1JKzQ3MmJLZE8reVB5aUhJc2d1NU8KZHhMaHRXdXBEeFRjOXNhVjlTN1dxR2xGUllyTnRxUmpUUnM2cTI1VmFkbkp6UDVURFRISlRaT1N0MWxBaTBIMwo3TSsvbllLcTNvL1o1Q3NZZUtxS0NtVDdrYUZvcTBXWm5tOVdqcVc4L09ReEo4eXlNRWJIbC93ZzhFcmp5T0JMCjljWng2ZmdONXZwd0V5R3REYVMyZXFmWmNYcjhWbmo5UzJPVnVoaklJTEZCV2hUaXk3dSs5YTlSeXJXWVRmQVUKNVlXb0tZRWNNb2hNWk9TdCsyMWkrQ3FOR3UxOGFBK0tvTi9wOXAxa2tBM2U5L0xMQjVxalRoNTV4ZEtqU201OApsa0FtK0lGOGZjeVE4d0tDQVFFQS8zWGJTb0RkV09Pa0NOaTRuU0QwakhXYVRCdUVyZFRaRGp2MFdSMGhuOUxFCk80V09NS2R4MkZoRUVUTlRTRnlCbDArWVhMUFp5TUhQWGRPcDJNU1hHRjJMa3B0Unphc2xabE92d3c5M3NtQWQKalh4THQ5c3BJMDc4aUlDaGtrRzI1YmZETHYwNGxsaURhT0hnbGZIb2hzSGVVMm5meDFtNXRaSkw4VDZ0L2w2OAozTnVEZ2J6VEJUUWxDNW0zZWxKUnJBMGtubkdVYkhuSGl1RXYvUkV0VFFJK2xRQzZVeDdNVzhLakZLcnZEYmxlCjJTMVo0aXVNdTJyRThxdTZXMnFLU0VHYUtVYXQrZFFUaTNSWGFHQjByeVU0Njl3ckxHWEtKWU9tSVVNTWtLRVUKcm5ldDRMVGxlVUl6LytmN0hWaVRJblJKbUh0eHY3T09wOWFWQkFDRTNRS0NBUUJhVUNlbk0wSXhsanNxRUxPRgpXdEV1NHhMN0owUHJZR3RkOVEwZXNjSXEzSktxTmxBbHc0UWlKNWV1ZktlZ1BXUlI0V05FRWpiMUNheXh2amVmCmVYK254YTBWd2tlMlladjJha3EyK0xET3FwL0V5dmRFcCtYSUx3aVlyYTRxeWNvemRDVVRvc3RPOUsyeFZhSzIKV1JyRTEzcHEwQ1JIdlNOL3dHOHFzY29HdUptUzBRdlhaTkYvaVh3TVRoZzNZYmdPcU9FYWY2c25ybUZ4dU81NApiYjRVMjVIb2JvZW1xVFBpcjhjTEJ1dW14K1JGd1NNNEFhV0FaN1ZGTHNCK0F2TTBBbVpqUGZJZzYyRW1IRS91ClRrYnBrbFhOYzBUTEErZHFaVlR6SmpWZldkdUg0OUpNZllGZnZ2eVdhUXYxd0o5dGxDM3BBcWdsL0FJR1hKMjEKV2wrNUFvSUJBUUNBYWc1QjlXVlBPWjlMa3lXNlZzYVhsTVpyVzZubWVBaGtUMThWTEwrcVI1aTlFdlhYSDFUcgpHL09iY2VGN2pvVFdPdnlzSFB4dEhuL08rb2Qybjl3VU0zSGwwSTNlaFJuWHhsbWtCREozU0RtUzFkaDZJNEVhCmJxb2pOQVlFTDJSeE5WNjRVRElSNFl5dFBGNVpBYkdSakFMbWkySWlJOGhtclpQZlFWTFZlMm9VMjRQUHc4dVAKNGhaOUxsZzg5UVI1S0dJd0E0aEl2eTYvY1ZOZU16UkNjYzVqdUdueFBwb2J4a3VVS0JqakhUdkFrRHVPcGJXbwp3Sk1nS1pNVVFRb2lLRmdHUjdGTTFINitxMDBFK3ZlczAxazV0d004M2k0dmpXNEZhbDcybFFLQWNVakduQ2grCk8vbEh1cVZRcXlYTFBBbUN6UXJ4QXFIRFhPcXJpRmRCQW9JQkFRQ0FGU0NtZWRsOHVZUEdKUGNhQlovWkNPR3QKemRvcU5qN2k5akVkZ1VRRWpHVHVKVDljUmZpaFI0cGpFaHFRdTRyNFhRcDBqY21adXllUnU2MEV4VExDSVdjNwpSc2pUeHU2VTZaWllNMDJST29mNlVKenREMkpBRURqRDhDZm85T2RaYWdhODRhbVNlSkpYd0VEQnVTY2MrL0gvCkh1cEViMnRDZjdqR3pCRlYwNVJlUVpRQlo3YkJhdmM2TzlPNWhkRDIxYjlwS2xFWDBVYXQyYW1LNENETjNmSDcKbTRlQ1AxRnVDNnJCV0RVSEhZaURBeGVjRExCeTFIRjZJeE9sSXBXekZZUFdJNGp6REhkMmhLMUpVS0QwSGR3ego5QjgxR2prY280dmZtbmxUMWFMS1IyVXZSK1Z2U2pmUjhCMk5xK0w5bkgxU3k0SndNaGEwWEc4ZHJrTEwKLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"),
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hwameistor-admission-ca",
			Namespace: "hwameistor",
		},
		Immutable:  nil,
		Data:       secret_data,
		StringData: nil,
		Type:       "Opaque",
	}
	err = client.Create(ctx, secret)
	if err != nil {
		logrus.Error("%+v ", err)
		f.ExpectNoError(err)
	}
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
	time.Sleep(3 * time.Minute)
	err = wait.PollImmediate(3*time.Second, 15*time.Minute, func() (done bool, err error) {
		err = client.Get(ctx, drbdKey1, drbd1)
		if err != nil {
			logrus.Error(err)
			f.ExpectNoError(err)
		}
		err = client.Get(ctx, drbdKey2, drbd2)
		if err != nil {
			logrus.Error(err)
			f.ExpectNoError(err)
		}
		if drbd1.Status.Succeeded == int32(1) && drbd2.Status.Succeeded == int32(1) {
			logrus.Printf("drbd ready")
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
		output, _ := RunInLinux("kubectl get pod -A  |grep -v Running |wc -l")
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
	err = installHwameiStorByHelm()
	if err != nil {
		logrus.Printf(" installHwameiStorByHelm_offline ERROR:%+v ", err)
		return false
	}
	addLabels()
	f := framework.NewDefaultFramework(v1alpha1.AddToScheme)
	client := f.GetClient()

	secret_data := map[string][]byte{
		"tls.crt": String2Bytes("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZzVENDQTVtZ0F3SUJBZ0lDQm5vd0RRWUpLb1pJaHZjTkFRRUxCUUF3R0RFV01CUUdBMVVFQ2hNTmFIZGgKYldWcGMzUnZjaTVwYnpBZUZ3MHlOREEwTVRVd056UTFNRE5hRncweU5UQTBNVFV3TnpRMU1ETmFNRkV4RmpBVQpCZ05WQkFvVERXaDNZVzFsYVhOMGIzSXVhVzh4TnpBMUJnTlZCQU1UTG1oM1lXMWxhWE4wYjNJdFlXUnRhWE56CmFXOXVMV052Ym5SeWIyeHNaWEl1YUhkaGJXVnBjM1J2Y2k1emRtTXdnZ0lpTUEwR0NTcUdTSWIzRFFFQkFRVUEKQTRJQ0R3QXdnZ0lLQW9JQ0FRRE9mTmlqWnErWTNlSHF2Z0VpOGpLdGtsSnp2RjM0YzFqUVQ5YnFBM3RnUjIvNApVbWQ0dmJKcFIzUUhZL1RRYjMwZS93SVp5OTdyNVA4SmJkS1RNWmMrdEI0WTJtK1c2YXJhMmd5UDBUQzlHTHpOCis5VzNlTnhheWh3Y1VTK08rQTc2Sll0MkVwcVp3TFd4bjA1bHc0OW16STVkbFNnMXV2WTcvYlhYQjFQVFlDajEKTVk2c2VhNXI3TitHSFluSG1GSEJTWi96UE02TEhsaUQ5ay9hQm4wVXRrR09TYkpxc0xBUFptOWg4SXFveDNYUwpEQTRWOG1rb2xTNDliMW5QL0RLekpGdFNySllaK2lyYXl2NU9aWURvblRVMCs5b09IbFFZNjZvNUVlellzZGtVCkhPUXQzazRLWTJEUERkNy9TVWt0ak9SL2ZYdGVBeEIwTDhuSi9xaHREOXdJQnFGbDVsQ1ZGeW51RlJKTVJMekYKVDVvcVluZGxuTjBLMWxTc3lBUzNkV2R3T25vdzhrVGlKbVRPTE1VdU5pV2tLcktmdnpzMENENGlsTUp6YU1sQwpBU0lSOU43eUZrN0FSTHREWlRSa2M1SmhCOStlR1pnYjdnd29CZ2FRTm43QmlNcmNxZG5JL0FDT3IwSmlCOFVYCkpBQUpWQ1kySDlja0tHUjZ6dGpYUzh0cTlHYnR1OTZzbGdpN05yakUvaElvQmZyZllJMkFWSDExSHZBU2dlMjcKMFkycnFqSXBrTVdxQzBSVUdleVV0alpld2EvUGFWU0Y0N0xpY0E0WlI1TDVadGxiRnNFSStxbEgxb3FONmNKSQpuL2YvclNaeVBpZmJGZmRLL1YxTWdQeHl3ajEyKzkwQlFsdjYyYXdWTmdJNmpPTHpTZUJMQTZzSTRGWnR4d0lECkFRQUJvNEhMTUlISU1BNEdBMVVkRHdFQi93UUVBd0lIZ0RBZEJnTlZIU1VFRmpBVUJnZ3JCZ0VGQlFjREFnWUkKS3dZQkJRVUhBd0V3RGdZRFZSME9CQWNFQlFFQ0F3UUdNSUdHQmdOVkhSRUVmekI5Z2g5b2QyRnRaV2x6ZEc5eQpMV0ZrYldsemMybHZiaTFqYjI1MGNtOXNiR1Z5Z2lwb2QyRnRaV2x6ZEc5eUxXRmtiV2x6YzJsdmJpMWpiMjUwCmNtOXNiR1Z5TG1oM1lXMWxhWE4wYjNLQ0xtaDNZVzFsYVhOMGIzSXRZV1J0YVhOemFXOXVMV052Ym5SeWIyeHMKWlhJdWFIZGhiV1ZwYzNSdmNpNXpkbU13RFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUJjMTdWd0NwY3RBYXBLOAptVHNuWlNtOWw5OC93TUp6eGhGRlpYUTNVVjBaRkozZ01SYThwZ2NZcTZ6bGl4YWNjdWJQeHhEWmtWNUwwVmJpCllwN1VuTEVITlRyLytaQTlnVDZGWDU2ekIxTFhMcFI5SVRwVzB1Z1ZKeUtrK3ExOWJrektkdzdGN0szT2lvU3MKT09YNm5WZHFqVHhkTkE2TVA5b2ZXN08xYXBTY1FWMTYwUmoyUXpINW9BZFJncUFkNEduVTkzaGQ5VmtJU0FRZwpUQ21MSWZwYjJHbVQrTzUxTVJsckVRSzl6dVJKSXRQS28zalNkUi9pUlVBdVBzZTNmaldHZmZtYU43b3B2bWVwCllqZWpjMFpTZHpEQTlkTDRwSW9PSWdVQUxDb1Y5aU4xSkhLSDl4NnJkMnFQdjRURUFEN25LcXNxNmVKbXVrQ2QKeEJjQUN4SkpYTkplODUwSDd5cDRxWWQxbjRNSEE1emtCWlZPN2htWHUxT2pOMFJrWFdyUGtNZDh4NDZjbnU4bApnWENNWXU2VEUwZFpEVVpFbjBwV25qL29tYWNuZ2VUZGpLcjAxMmVDR0ZLSjdYT1ladWRCcEdKcWtHNHZSVlBHCnAwdGJoU3dVeFJNQzlQcDc5cGJncnhjbXJZdHpqcjR0RU12Qk56azI3QUlNQVRSb0Y5YmxLVzdLNWtsdmNoc3UKMitSOXhsbkVPcHlXdUsrUTJ6ZVc0Vm5uY3ZDaEFYdU1Kc1hSeHNlOXlUbjh4OWdieEJqRlJTcFZXV2dUZGl1cQpSUlQ4NHNDbmRkbjRWSHl4eUMraHpOUk1FalNGYUZpSm16UnZHVzdjbmdEZkcrbGVwRlNaVFpyS2NYa3B4TXBvCkdlbzU2b3o1d2lUTzZ3c0lwblBxTHZaeU5ISTkKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="),
		"tls.key": String2Bytes("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKS1FJQkFBS0NBZ0VBem56WW8yYXZtTjNoNnI0Qkl2SXlyWkpTYzd4ZCtITlkwRS9XNmdON1lFZHYrRkpuCmVMMnlhVWQwQjJQMDBHOTlIdjhDR2N2ZTYrVC9DVzNTa3pHWFByUWVHTnB2bHVtcTJ0b01qOUV3dlJpOHpmdlYKdDNqY1dzb2NIRkV2anZnTytpV0xkaEthbWNDMXNaOU9aY09QWnN5T1haVW9OYnIyTy8yMTF3ZFQwMkFvOVRHTwpySG11YSt6ZmhoMkp4NWhSd1VtZjh6ek9peDVZZy9aUDJnWjlGTFpCamtteWFyQ3dEMlp2WWZDS3FNZDEwZ3dPCkZmSnBLSlV1UFc5Wnovd3lzeVJiVXF5V0dmb3Eyc3IrVG1XQTZKMDFOUHZhRGg1VUdPdXFPUkhzMkxIWkZCemsKTGQ1T0NtTmd6dzNlLzBsSkxZemtmMzE3WGdNUWRDL0p5ZjZvYlEvY0NBYWhaZVpRbFJjcDdoVVNURVM4eFUrYQpLbUozWlp6ZEN0WlVyTWdFdDNWbmNEcDZNUEpFNGlaa3ppekZMallscENxeW43ODdOQWcrSXBUQ2MyakpRZ0VpCkVmVGU4aFpPd0VTN1EyVTBaSE9TWVFmZm5obVlHKzRNS0FZR2tEWit3WWpLM0tuWnlQd0FqcTlDWWdmRkZ5UUEKQ1ZRbU5oL1hKQ2hrZXM3WTEwdkxhdlJtN2J2ZXJKWUl1emE0eFA0U0tBWDYzMkNOZ0ZSOWRSN3dFb0h0dTlHTgpxNm95S1pERnFndEVWQm5zbExZMlhzR3Z6MmxVaGVPeTRuQU9HVWVTK1diWld4YkJDUHFwUjlhS2plbkNTSi8zCi82MG1jajRuMnhYM1N2MWRUSUQ4Y3NJOWR2dmRBVUpiK3Rtc0ZUWUNPb3ppODBuZ1N3T3JDT0JXYmNjQ0F3RUEKQVFLQ0FnQlRBdmZwQVFBNUdZS1laNWgzWHgwbWtIdWJVYkFKa1AyMGlWUmlNZElveTY5bENtUFVzNWFGSXBldgozT3h2MjFKVzRyZ0tmTjU2d1U0U2tWY2xtN3N0RnlZK1lacVk5Z21kQUNvcXF6UEg0YWtOdWwrWjZmblVlZ3BCCkRCWnRpS1lreEhGRG40a05wTjZtc04vSW80MEVDOXAwSHVIU2IvNUw4ZHhudk9kalEvWlRWSUFDTU0wQlVISXcKTGJIVlR4UVcrNjFWRVM0TDJjQkdGUURYS3VLZ01BRTgzYjNmakhiTFNpZ2ZISi94NkhZaVlueExNSEE4UTNVYQpkNmlJcCtBT0N1QzM1UEJYTXk2RlZUUlZlaE1QOFJ1UjJHbzZZTkQ0VDgzWU1hK2hFaWF1SWt0TmNXQzRabWxsCjlES2hPT28xYUwzM3p3K0lwUHk1YVBqYnZ3Z2dUbXRHWDhldTJ5K3NiMWxOSUZzK0RzYzY0WDVuWFMxR2drTzUKRWEwNngvM0E1ajRVYWVKRXdvSnVRNkNBdEwvUGUrN2N0d1hBdExRV2s3M1k2YStsOEhrbERBTmV1SDFkWVZvTwpVQ2NEUGxuZFk3dEt0NUtSeXdEcUU5N0gzeEhTU3g1R1AwSUJqSEN6NVNDNlg2Vk94YlJSbEdjQ2NsOS8raGdDCllSNWFGcFVqV0FrWTlzNEE3bEdhVXl0M3VrQVFMSCt4OUl3R0wyQUx6Ujdzb1QxMUpZZEhyWVBwSGZJUFA4MXkKVUJ1VytubTk3QzNlamV5d1RPNTMvRkJMeTRqdkRqWk9ETXEySFJmbmhjMGMxbmhlcTlEN1lFNnBEL3Nna0wwagpoazgwaFJGbEw3SGVGQU9hd1lsM2xBY01WTFgzUjZDMFFVLytRTDNzOU1oRzhIZVFxUUtDQVFFQXp1eUJ6VldiCnUrbnJLdUxmcCt0TGhBZWJMSEpuNmJNOFBFb2h3YUMxMUdLajh3eW9nZW1JKzQ3MmJLZE8reVB5aUhJc2d1NU8KZHhMaHRXdXBEeFRjOXNhVjlTN1dxR2xGUllyTnRxUmpUUnM2cTI1VmFkbkp6UDVURFRISlRaT1N0MWxBaTBIMwo3TSsvbllLcTNvL1o1Q3NZZUtxS0NtVDdrYUZvcTBXWm5tOVdqcVc4L09ReEo4eXlNRWJIbC93ZzhFcmp5T0JMCjljWng2ZmdONXZwd0V5R3REYVMyZXFmWmNYcjhWbmo5UzJPVnVoaklJTEZCV2hUaXk3dSs5YTlSeXJXWVRmQVUKNVlXb0tZRWNNb2hNWk9TdCsyMWkrQ3FOR3UxOGFBK0tvTi9wOXAxa2tBM2U5L0xMQjVxalRoNTV4ZEtqU201OApsa0FtK0lGOGZjeVE4d0tDQVFFQS8zWGJTb0RkV09Pa0NOaTRuU0QwakhXYVRCdUVyZFRaRGp2MFdSMGhuOUxFCk80V09NS2R4MkZoRUVUTlRTRnlCbDArWVhMUFp5TUhQWGRPcDJNU1hHRjJMa3B0Unphc2xabE92d3c5M3NtQWQKalh4THQ5c3BJMDc4aUlDaGtrRzI1YmZETHYwNGxsaURhT0hnbGZIb2hzSGVVMm5meDFtNXRaSkw4VDZ0L2w2OAozTnVEZ2J6VEJUUWxDNW0zZWxKUnJBMGtubkdVYkhuSGl1RXYvUkV0VFFJK2xRQzZVeDdNVzhLakZLcnZEYmxlCjJTMVo0aXVNdTJyRThxdTZXMnFLU0VHYUtVYXQrZFFUaTNSWGFHQjByeVU0Njl3ckxHWEtKWU9tSVVNTWtLRVUKcm5ldDRMVGxlVUl6LytmN0hWaVRJblJKbUh0eHY3T09wOWFWQkFDRTNRS0NBUUJhVUNlbk0wSXhsanNxRUxPRgpXdEV1NHhMN0owUHJZR3RkOVEwZXNjSXEzSktxTmxBbHc0UWlKNWV1ZktlZ1BXUlI0V05FRWpiMUNheXh2amVmCmVYK254YTBWd2tlMlladjJha3EyK0xET3FwL0V5dmRFcCtYSUx3aVlyYTRxeWNvemRDVVRvc3RPOUsyeFZhSzIKV1JyRTEzcHEwQ1JIdlNOL3dHOHFzY29HdUptUzBRdlhaTkYvaVh3TVRoZzNZYmdPcU9FYWY2c25ybUZ4dU81NApiYjRVMjVIb2JvZW1xVFBpcjhjTEJ1dW14K1JGd1NNNEFhV0FaN1ZGTHNCK0F2TTBBbVpqUGZJZzYyRW1IRS91ClRrYnBrbFhOYzBUTEErZHFaVlR6SmpWZldkdUg0OUpNZllGZnZ2eVdhUXYxd0o5dGxDM3BBcWdsL0FJR1hKMjEKV2wrNUFvSUJBUUNBYWc1QjlXVlBPWjlMa3lXNlZzYVhsTVpyVzZubWVBaGtUMThWTEwrcVI1aTlFdlhYSDFUcgpHL09iY2VGN2pvVFdPdnlzSFB4dEhuL08rb2Qybjl3VU0zSGwwSTNlaFJuWHhsbWtCREozU0RtUzFkaDZJNEVhCmJxb2pOQVlFTDJSeE5WNjRVRElSNFl5dFBGNVpBYkdSakFMbWkySWlJOGhtclpQZlFWTFZlMm9VMjRQUHc4dVAKNGhaOUxsZzg5UVI1S0dJd0E0aEl2eTYvY1ZOZU16UkNjYzVqdUdueFBwb2J4a3VVS0JqakhUdkFrRHVPcGJXbwp3Sk1nS1pNVVFRb2lLRmdHUjdGTTFINitxMDBFK3ZlczAxazV0d004M2k0dmpXNEZhbDcybFFLQWNVakduQ2grCk8vbEh1cVZRcXlYTFBBbUN6UXJ4QXFIRFhPcXJpRmRCQW9JQkFRQ0FGU0NtZWRsOHVZUEdKUGNhQlovWkNPR3QKemRvcU5qN2k5akVkZ1VRRWpHVHVKVDljUmZpaFI0cGpFaHFRdTRyNFhRcDBqY21adXllUnU2MEV4VExDSVdjNwpSc2pUeHU2VTZaWllNMDJST29mNlVKenREMkpBRURqRDhDZm85T2RaYWdhODRhbVNlSkpYd0VEQnVTY2MrL0gvCkh1cEViMnRDZjdqR3pCRlYwNVJlUVpRQlo3YkJhdmM2TzlPNWhkRDIxYjlwS2xFWDBVYXQyYW1LNENETjNmSDcKbTRlQ1AxRnVDNnJCV0RVSEhZaURBeGVjRExCeTFIRjZJeE9sSXBXekZZUFdJNGp6REhkMmhLMUpVS0QwSGR3ego5QjgxR2prY280dmZtbmxUMWFMS1IyVXZSK1Z2U2pmUjhCMk5xK0w5bkgxU3k0SndNaGEwWEc4ZHJrTEwKLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K"),
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hwameistor-admission-ca",
			Namespace: "hwameistor",
		},
		Immutable:  nil,
		Data:       secret_data,
		StringData: nil,
		Type:       "Opaque",
	}
	err = client.Create(ctx, secret)
	if err != nil {
		logrus.Error("%+v ", err)
		f.ExpectNoError(err)
	}

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
	_, _ = RunInLinux("helm list -A | grep 'hwameistor' | awk '{print $1}' | xargs helm uninstall -n hwameistor")
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
				Owner:    "local-storage",
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

	err := wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
		for _, nodes := range nodelist.Items {
			time.Sleep(3 * time.Second)
			localDiskClaim := &v1alpha1.LocalDiskClaim{}
			localDiskClaimKey := k8sclient.ObjectKey{
				Name:      "localdiskclaim-" + nodes.Name,
				Namespace: "kube-system",
			}
			err := client.Get(ctx, localDiskClaimKey, localDiskClaim)
			if !k8serror.IsNotFound(err) {
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

func CreateLdcForLdm(ctx context.Context) error {
	logrus.Printf("create ldc for each node")
	nodelist := nodeList()
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()

	LdList := &v1alpha1.LocalDiskList{}
	err := client.List(ctx, LdList)
	for _, ld := range LdList.Items {
		logrus.Printf(ld.Spec.Owner)
		if ld.Spec.Owner == "" {
			ld.Spec.Owner = "local-disk-manager"
			err := client.Update(ctx, &ld)
			if err != nil {
				logrus.Printf("Update LDC failed ：%+v ", err)
				f.ExpectNoError(err)
			}
		}
	}

	for _, nodes := range nodelist.Items {
		exmlocalDiskClaim := &v1alpha1.LocalDiskClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "localdiskclaim-" + nodes.Name,
				Namespace: "kube-system",
			},
			Spec: v1alpha1.LocalDiskClaimSpec{
				Owner:    "local-disk-manager",
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

	err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
		for _, nodes := range nodelist.Items {
			time.Sleep(3 * time.Second)
			localDiskClaim := &v1alpha1.LocalDiskClaim{}
			localDiskClaimKey := k8sclient.ObjectKey{
				Name:      "localdiskclaim-" + nodes.Name,
				Namespace: "kube-system",
			}
			err := client.Get(ctx, localDiskClaimKey, localDiskClaim)
			if !k8serror.IsNotFound(err) {
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
	pvcList := &corev1.PersistentVolumeClaimList{}
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

	err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
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
	err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
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
		&corev1.PodExecOptions{
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
	_, _ = RunInLinux("sh install_drbd.sh")

}

//Get the corresponding pod by deploy
func GetPodsByDeploy(ctx context.Context, namespace, deployName string) (*corev1.PodList, error) {
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	deploy := &appsv1.Deployment{}
	deployKey := k8sclient.ObjectKey{
		Name:      deployName,
		Namespace: namespace,
	}
	err := client.Get(ctx, deployKey, deploy)
	if err != nil {
		logrus.Error("get deploy error", err)
		f.ExpectNoError(err)
	}
	podList := &corev1.PodList{}
	err = client.List(ctx, podList, k8sclient.InNamespace(deploy.Namespace), k8sclient.MatchingLabels(deploy.Spec.Selector.MatchLabels))
	if err != nil {
		logrus.Error("get pod list error", err)
		f.ExpectNoError(err)
	}
	return podList, nil
}

//Output the events of the target podlist
func GetPodEvents(ctx context.Context, podList *corev1.PodList) {
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	for _, pod := range podList.Items {
		eventList := &corev1.EventList{}
		err := client.List(ctx, eventList, k8sclient.InNamespace(pod.Namespace), k8sclient.MatchingFields{"involvedObject.name": pod.Name})
		if err != nil {
			logrus.Error("get event list error", err)
			f.ExpectNoError(err)
		}
		for _, event := range eventList.Items {
			logrus.Printf("event:%+v", event)
		}
	}
}

//Output the events of all pods under the default namespace
func GetAllPodEventsInDefaultNamespace(ctx context.Context) {
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	podList := &corev1.PodList{}
	err := client.List(ctx, podList, k8sclient.InNamespace("default"))
	if err != nil {
		logrus.Error("get pod list error", err)
		f.ExpectNoError(err)
	}
	logrus.Printf("Output the events of all pods under the default namespace")
	for _, pod := range podList.Items {
		eventList := &corev1.EventList{}
		err := client.List(ctx, eventList, k8sclient.InNamespace(pod.Namespace), k8sclient.MatchingFields{"involvedObject.name": pod.Name})
		if err != nil {
			logrus.Error("get event list error", err)
			f.ExpectNoError(err)
		}
		for _, event := range eventList.Items {
			logrus.Printf("event-Reason:%+v", event.Reason)
			logrus.Printf("event-Message:%+v", event.Message)
		}
	}

}

//return All Pod In Hwameistor Namespace
func GetAllPodInHwameistorNamespace(ctx context.Context) *corev1.PodList {
	f := framework.NewDefaultFramework(clientset.AddToScheme)
	client := f.GetClient()
	podList := &corev1.PodList{}
	err := client.List(ctx, podList, k8sclient.InNamespace("hwameistor"))
	if err != nil {
		logrus.Error("get pod list error", err)
		f.ExpectNoError(err)
	}
	return podList
}

//Get logs of target pod
func getPodLogs(pod corev1.Pod) {
	podLogOpts := corev1.PodLogOptions{}
	config, err := config.GetConfig()
	if err != nil {
		logrus.Error("error in getting config")
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Error("error in getting access to K8S")
	}
	// 循环输出这个pod的每个container
	for _, container := range pod.Spec.Containers {
		logrus.Printf("container name:%+v", container.Name)
		podLogOpts.Container = container.Name
		req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
		ctx := context.TODO()
		podLogs, err := req.Stream(ctx)
		if err != nil {
			logrus.Error("error in opening stream")
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			logrus.Error("error in copy information from podLogs to buf")
		}
		str := buf.String()
		if strings.Contains(str, "error") {
			logrus.Infoln(str)
		}

	}

}

//return All Pod logs In Hwameistor Namespace
func GetAllPodLogsInHwameistorNamespace(ctx context.Context) {
	podList := GetAllPodInHwameistorNamespace(ctx)
	for _, pod := range podList.Items {
		logrus.Printf("pod:%+v", pod.Name)
		getPodLogs(pod)

	}
}

func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
