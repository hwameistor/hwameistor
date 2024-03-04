package E2eTest

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
)

var _ = ginkgo.Describe("test convertible localstorage volume ", ginkgo.Label("periodCheck"), func() {

	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", ginkgo.FlakeAttempts(5), func() {
		result := utils.ConfigureEnvironment(ctx)
		gomega.Expect(result).To(gomega.BeNil())
		f := framework.NewDefaultFramework(clientset.AddToScheme)
		client = f.GetClient()
		utils.CreateLdc(ctx)

	})

	ginkgo.Context("create a StorageClass", func() {
		ginkgo.It("create a sc", func() {
			//create sc
			deleteObj := corev1.PersistentVolumeReclaimDelete
			waitForFirstConsumerObj := storagev1.VolumeBindingWaitForFirstConsumer
			examplesc := &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "local-storage-hdd-lvm",
				},
				Provisioner: "lvm.hwameistor.io",
				Parameters: map[string]string{
					"replicaNumber":             "1",
					"convertible":               "true",
					"poolClass":                 "HDD",
					"csi.storage.k8s.io/fstype": "xfs",
				},
				ReclaimPolicy:        &deleteObj,
				AllowVolumeExpansion: utils.BoolPter(true),
				VolumeBindingMode:    &waitForFirstConsumerObj,
			}
			err := client.Create(ctx, examplesc)
			if err != nil {
				logrus.Printf("Create SC failed ：%+v ", err)
				f.ExpectNoError(err)
			}
		})
	})
	ginkgo.Context("create a PVC", func() {
		ginkgo.It("create PVC", func() {
			//create PVC
			storageClassName := "local-storage-hdd-lvm"
			examplePvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-lvm",
					Namespace: "default",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					StorageClassName: &storageClassName,
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			}
			err := client.Create(ctx, examplePvc)
			if err != nil {
				logrus.Printf("Create PVC failed ：%+v ", err)
				f.ExpectNoError(err)
			}

			gomega.Expect(err).To(gomega.BeNil())
		})

	})
	ginkgo.Context("create a deployment", func() {

		ginkgo.It("create deployment", func() {
			//create deployment
			exampleDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.DeploymentName,
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: utils.Int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "demo",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "demo",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "web",
									Image: "10.6.112.210/hwameistor/dao-2048:v1.2.0",
									Ports: []corev1.ContainerPort{
										{
											Name:          "http",
											Protocol:      corev1.ProtocolTCP,
											ContainerPort: 80,
										},
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "2048-volume-lvm",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "2048-volume-lvm",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "pvc-lvm",
										},
									},
								},
							},
						},
					},
				},
			}
			err := client.Create(ctx, exampleDeployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("PVC STATUS should be Bound", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			pvcKey := ctrlclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			logrus.Infof("Waiting for the PVC to be bound")
			err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err = client.Get(ctx, pvcKey, pvc); pvc.Status.Phase != corev1.ClaimBound {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("PVC binding timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("deploy STATUS should be AVAILABLE", func() {
			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			logrus.Infof("waiting for the deployment to be ready ")
			err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err = client.Get(ctx, deployKey, deployment); deployment.Status.AvailableReplicas != int32(1) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("deployment ready timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})

	})
	ginkgo.Context("Test the volume", func() {

		ginkgo.It("write test data", func() {

			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName,
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"demo"})
			selector := labels.NewSelector()
			selector = selector.Add(*apps)
			listOption := ctrlclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &corev1.PodList{}
			err = client.List(ctx, podlist, &listOption)

			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			containers := deployment.Spec.Template.Spec.Containers
			for _, pod := range podlist.Items {
				for _, container := range containers {
					_, _, err := utils.ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && echo it-is-a-test >test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					output, _, err := utils.ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && cat test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					gomega.Expect(output).To(gomega.Equal("it-is-a-test"))
				}
			}
		})
		ginkgo.It("convert volume", func() {
			lvlist := &v1alpha1.LocalVolumeList{}
			err := client.List(ctx, lvlist)
			if err != nil {
				logrus.Error("%+v ", err)
				f.ExpectNoError(err)
			}
			lvname := lvlist.Items[0].Name
			lvc := &v1alpha1.LocalVolumeConvert{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "hwameistor.io/v1alpha1",
					Kind:       "LocalVolumeConvert",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localvolumeconvert-1",
					Namespace: "default",
				},
				Spec: v1alpha1.LocalVolumeConvertSpec{
					VolumeName:    lvname,
					ReplicaNumber: 2,
				},
			}
			err = client.Create(ctx, lvc)
			if err != nil {
				logrus.Printf("Create lvconvert failed ：%+v ", err)
				f.ExpectNoError(err)
			}

			err = wait.PollImmediate(10*time.Second, 10*time.Minute, func() (done bool, err error) {
				lvrlist := &v1alpha1.LocalVolumeReplicaList{}
				err = client.List(ctx, lvrlist)
				if err != nil {
					logrus.Error("%+v ", err)
					f.ExpectNoError(err)
				}
				if len(lvrlist.Items) != 2 {
					return false, nil
				} else {
					err := client.List(ctx, lvlist)
					if err != nil {
						logrus.Error("%+v ", err)
						f.ExpectNoError(err)
					}
					for _, lv := range lvlist.Items {
						if lv.Status.State != "Ready" {
							return false, nil
						}
					}
				}
				return true, nil

			})
			if err != nil {
				logrus.Printf("lv convert failed ：%+v ", err)
				f.ExpectNoError(err)
			}

		})
		ginkgo.It("Expand volume", func() {
			pvc := &corev1.PersistentVolumeClaim{}
			pvcKey := ctrlclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("Failed to find pvc ：%+v ", err)
				f.ExpectNoError(err)
			}
			storageMap := make(map[corev1.ResourceName]resource.Quantity)
			storageMap[corev1.ResourceStorage] = resource.MustParse("2Gi")
			pvc.Spec.Resources.Requests = storageMap
			err = client.Update(ctx, pvc)

			pvc = &corev1.PersistentVolumeClaim{}
			pvcKey = ctrlclient.ObjectKey{
				Name:      "pvc-lvm",
				Namespace: "default",
			}
			err = client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			logrus.Infof(pvc.Status.Capacity.Storage().String())
			logrus.Infof("Waiting for the PVC to be bound")
			err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err = client.Get(ctx, pvcKey, pvc); pvc.Status.Capacity.Storage().String() != "2Gi" {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("PVC binding timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})
		ginkgo.It("check test file", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName,
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"demo"})
			selector := labels.NewSelector()
			selector = selector.Add(*apps)
			listOption := ctrlclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &corev1.PodList{}
			err = client.List(ctx, podlist, &listOption)

			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			containers := deployment.Spec.Template.Spec.Containers
			for _, pod := range podlist.Items {
				for _, container := range containers {
					output, _, err := utils.ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && cat test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					gomega.Expect(output).To(gomega.Equal("it-is-a-test"))
				}
			}
		})
		ginkgo.It("Delete test data", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName,
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"demo"})
			selector := labels.NewSelector()
			selector = selector.Add(*apps)
			listOption := ctrlclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &corev1.PodList{}
			err = client.List(ctx, podlist, &listOption)

			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			containers := deployment.Spec.Template.Spec.Containers
			for _, pod := range podlist.Items {
				for _, container := range containers {
					_, _, err := utils.ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && rm -rf test", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					output, _, err := utils.ExecInPod(config, deployment.Namespace, pod.Name, "cd /data && ls", container.Name)
					if err != nil {
						logrus.Printf("%+v ", err)
						f.ExpectNoError(err)
					}
					gomega.Expect(output).To(gomega.Equal(""))
				}
			}
		})
	})
	ginkgo.Context("check the localvolume", func() {

		ginkgo.It("check the accessibility node", func() {

			lvlist := &v1alpha1.LocalVolumeList{}
			err := client.List(ctx, lvlist)
			if err != nil {
				logrus.Error("%+v ", err)
				f.ExpectNoError(err)
			}
			lv := lvlist.Items[0]
			logrus.Printf("check the lv")
			for _, replicas := range lv.Spec.Config.Replicas {
				result := true
				for _, node := range lv.Spec.Accessibility.Nodes {
					if replicas.Hostname == node {
						result = true
						break
					}
				}
				gomega.Expect(result).To(gomega.BeTrue())
			}

		})

	})
	ginkgo.Context("Clean up the environment", func() {
		ginkgo.It("Delete test Deployment", func() {
			//delete deploy
			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			logrus.Infof("deleting test Deployment ")

			err = client.Delete(ctx, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			err = wait.PollImmediate(3*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err := client.Get(ctx, deployKey, deployment); !k8serror.IsNotFound(err) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())

		})
		ginkgo.It("delete all pvc ", func() {
			err := utils.DeleteAllPVC(ctx)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("delete all sc", func() {
			err := utils.DeleteAllSC(ctx)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("delete helm", func() {
			utils.UninstallHelm()
		})
	})

})
