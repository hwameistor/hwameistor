package E2eTest

import (
	"context"
	clientset "github.com/hwameistor/hwameistor/pkg/apis/client/clientset/versioned/scheme"
	v1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

var _ = ginkgo.Describe("test localstorage Ha volume", ginkgo.Label("periodCheck"), func() {

	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", func() {
		result := utils.ConfigureEnvironment(ctx)
		gomega.Expect(result).To(gomega.BeNil())
		f = framework.NewDefaultFramework(clientset.AddToScheme)
		client = f.GetClient()
		utils.CreateLdc(ctx)
	})
	ginkgo.Context("create a HA-StorageClass", func() {
		ginkgo.It("create a sc", func() {
			//create sc
			deleteObj := apiv1.PersistentVolumeReclaimDelete
			waitForFirstConsumerObj := storagev1.VolumeBindingWaitForFirstConsumer
			examplesc := &storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "local-storage-hdd-lvm-ha",
				},
				Provisioner: "lvm.hwameistor.io",
				Parameters: map[string]string{
					"replicaNumber":             "2",
					"poolClass":                 "HDD",
					"poolType":                  "REGULAR",
					"volumeKind":                "LVM",
					"striped":                   "true",
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
	ginkgo.Context("create a HA-PersistentVolumeClaim", func() {
		ginkgo.It("create PVC", func() {
			//create PVC
			storageClassName := "local-storage-hdd-lvm-ha"
			examplePvc := &apiv1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-lvm-ha",
					Namespace: "default",
				},
				Spec: apiv1.PersistentVolumeClaimSpec{
					AccessModes:      []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
					StorageClassName: &storageClassName,
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceStorage: resource.MustParse("100Mi"),
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
			exampleDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.HaDeploymentName,
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
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "demo",
							},
						},
						Spec: apiv1.PodSpec{
							SchedulerName: "hwameistor-scheduler",
							Affinity: &apiv1.Affinity{
								NodeAffinity: &apiv1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
										NodeSelectorTerms: []apiv1.NodeSelectorTerm{
											{
												[]apiv1.NodeSelectorRequirement{
													{
														Key:      "kubernetes.io/hostname",
														Operator: apiv1.NodeSelectorOpIn,
														Values: []string{
															"k8s-node1",
														},
													},
												},
												[]apiv1.NodeSelectorRequirement{},
											},
										},
									},
									PreferredDuringSchedulingIgnoredDuringExecution: nil,
								},
								PodAffinity:     nil,
								PodAntiAffinity: nil,
							},
							Containers: []apiv1.Container{
								{
									Name:  "web",
									Image: "ghcr.m.daocloud.io/daocloud/dao-2048:v1.2.0",
									Ports: []apiv1.ContainerPort{
										{
											Name:          "http",
											Protocol:      apiv1.ProtocolTCP,
											ContainerPort: 80,
										},
									},
									VolumeMounts: []apiv1.VolumeMount{
										{
											Name:      "2048-volume-lvm-ha",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []apiv1.Volume{
								{
									Name: "2048-volume-lvm-ha",
									VolumeSource: apiv1.VolumeSource{
										PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
											ClaimName: "pvc-lvm-ha",
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

			pvc := &apiv1.PersistentVolumeClaim{}
			pvcKey := k8sclient.ObjectKey{
				Name:      "pvc-lvm-ha",
				Namespace: "default",
			}
			err := client.Get(ctx, pvcKey, pvc)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			logrus.Infof("Waiting for the PVC to be bound")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, pvcKey, pvc); pvc.Status.Phase != apiv1.ClaimBound {
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
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			logrus.Infof("waiting for the deployment to be ready ")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
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
	ginkgo.Context("test volumes", func() {
		ginkgo.It("Write test file", func() {
			//create a request
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
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
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
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
		ginkgo.It("Delete test file", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
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
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
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
	ginkgo.Context("test HA-volumes", func() {
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
			for _, lvr := range lvrList.Items {
				if lvr.Spec.NodeName == "k8s-master" {
					exlvm := &v1alpha1.LocalVolumeMigrate{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "localvolumegroupmigrate-1",
							Namespace: "default",
						},
						Spec: v1alpha1.LocalVolumeMigrateSpec{
							TargetNodesSuggested: []string{"k8s-node2"},
							SourceNode:           "k8s-master",
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
					logrus.Infof("wait 3 minutes for migrate lv")
					time.Sleep(3 * time.Minute)
					break

				}

			}

		})

		ginkgo.It("Write test file", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
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
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
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
		ginkgo.It("update deploy", func() {
			//delete deploy
			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				f.ExpectNoError(err)
			}

			newAffinity := []apiv1.NodeSelectorTerm{
				{
					[]apiv1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/hostname",
							Operator: apiv1.NodeSelectorOpIn,
							Values: []string{
								"k8s-node2",
							},
						},
					},
					[]apiv1.NodeSelectorRequirement{},
				},
			}
			deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = newAffinity

			err = client.Update(ctx, deployment)
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}
			logrus.Infof("waiting for the deployment to be ready ")
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, deployKey, deployment); deployment.Status.AvailableReplicas != int32(1) {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Infof("deployment ready timeout")
				logrus.Error(err)
			}
		})
		ginkgo.It("check test file", func() {
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
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
			listOption := k8sclient.ListOptions{
				LabelSelector: selector,
			}
			podlist := &v1.PodList{}
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
			deployKey := k8sclient.ObjectKey{
				Name:      utils.HaDeploymentName,
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
			err = wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
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
