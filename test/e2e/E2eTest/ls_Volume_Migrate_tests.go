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

var _ = ginkgo.Describe("volume migrate test", ginkgo.Label("periodCheck"), func() {

	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", ginkgo.FlakeAttempts(5), func() {
		result := utils.ConfigureEnvironment(ctx)
		gomega.Expect(result).To(gomega.BeNil())
		f = framework.NewDefaultFramework(clientset.AddToScheme)
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
	ginkgo.Context("create a PersistentVolumeClaim", func() {
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
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Mi"),
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

		ginkgo.It("create a deployment", func() {
			//create deployment
			_, _ = utils.RunInLinux("kubectl taint node --all node-role.kubernetes.io/master-")
			exampleDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.DeploymentName,
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: utils.Int32Ptr(1),
					Strategy: appsv1.DeploymentStrategy{
						Type: appsv1.RecreateDeploymentStrategyType,
					},
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
											Name:      "2048-volume-lvm-ha",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "2048-volume-lvm-ha",
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
	ginkgo.Context("test volumes migrate", func() {
		SourceNode := ""
		ginkgo.It("Write test file", func() {
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
		ginkgo.It("Get the SourceNode", func() {

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

			logrus.Infof("The node where the application is located is: " + podlist.Items[0].Spec.NodeName)
			SourceNode = podlist.Items[0].Spec.NodeName
		})
		ginkgo.It("Stop the workload", func() {
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
			deployment.Spec.Replicas = utils.Int32Ptr(0)
			err = client.Update(ctx, deployment)
			logrus.Infof("waiting for the deployment to be stop")
			err = wait.PollImmediate(10*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err = client.Get(ctx, deployKey, deployment); deployment.Status.AvailableReplicas != int32(0) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("deployment stop timeout")
				logrus.Error(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("create a localvolumemigrate", func() {

			lvrList := &v1alpha1.LocalVolumeReplicaList{}
			err := client.List(ctx, lvrList)
			if err != nil {
				logrus.Printf("list lvr failed ：%+v ", err)
			}
			lvlist := &v1alpha1.LocalVolumeList{}
			err = client.List(ctx, lvlist)
			if err != nil {
				logrus.Error("%+v ", err)
				f.ExpectNoError(err)
			}

			lvname := lvlist.Items[0].Name
			exlvm := &v1alpha1.LocalVolumeMigrate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "localvolumegroupmigrate-1",
					Namespace: "default",
				},
				Spec: v1alpha1.LocalVolumeMigrateSpec{
					TargetNodesSuggested: []string{},
					SourceNode:           SourceNode,
					VolumeName:           lvname,
					MigrateAllVols:       true,
				},
			}

			err = client.Create(ctx, exlvm)
			logrus.Infof("create lvm")
			if err != nil {
				logrus.Printf("Create lvgm failed ：%+v ", err)
				f.ExpectNoError(err)
			}

		})
		ginkgo.It("check localvolumemigrate", func() {
			logrus.Infof("wait 20 minutes for migrate lv")
			err := wait.PollImmediate(10*time.Second, 20*time.Minute, func() (done bool, err error) {
				lmgList := &v1alpha1.LocalVolumeMigrateList{}
				err = client.List(ctx, lmgList)
				if err != nil {
					logrus.Printf("list lmg failed ：%+v ", err)
				}
				if len(lmgList.Items) != 0 {
					return false, nil
				} else {
					logrus.Info("migrate ready")
					return true, nil
				}

			})

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

			lvrList := &v1alpha1.LocalVolumeReplicaList{}
			err = client.List(ctx, lvrList)
			if err != nil {
				logrus.Printf("list lvr failed ：%+v ", err)
			}

			gomega.Expect(len(lvrList.Items)).To(gomega.Equal(1))
			for _, lvr := range lvrList.Items {
				gomega.Expect(SourceNode).To(gomega.Not(gomega.Equal(lvr.Spec.NodeName)))
			}

		})
		ginkgo.It("Start the workload", func() {
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
			deployment.Spec.Replicas = utils.Int32Ptr(1)
			err = client.Update(ctx, deployment)
			logrus.Infof("waiting for the workload to be start")
			err = wait.PollImmediate(10*time.Second, framework.PodStartTimeout, func() (done bool, err error) {
				if err = client.Get(ctx, deployKey, deployment); deployment.Status.AvailableReplicas != int32(1) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("workload start timeout")
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
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
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
		ginkgo.It("delete all pvc", func() {
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
