package E2eTest

import (
	"context"
	ldapis "github.com/hwameistor/hwameistor/pkg/apis/generated/local-disk-manager/clientset/versioned/scheme"
	lsv1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"time"
)

var _ = ginkgo.Describe("test localstorage Ha volume", ginkgo.Label("test"), func() {

	f := framework.NewDefaultFramework(lsv1.AddToScheme)
	client := f.GetClient()
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", func() {
		result := configureEnvironment(ctx)
		gomega.Expect(result).To(gomega.Equal(true))
		createLdc(ctx)
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
				AllowVolumeExpansion: boolPter(true),
				VolumeBindingMode:    &waitForFirstConsumerObj,
			}
			err := client.Create(ctx, examplesc)
			if err != nil {
				logrus.Printf("Create SC failed ：%+v ", err)
				f.ExpectNoError(err)
			}
		})
	})
	ginkgo.Context("Deduplication test", func() {
		for testNumbers := 1; testNumbers <= NumberOfDeduplicationTests; testNumbers++ {
			ginkgo.It(strconv.Itoa(testNumbers)+"th create PVC", func() {
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
			ginkgo.It(strconv.Itoa(testNumbers)+"th create a deployment", func() {
				//create deployment
				exampleDeployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      HaDeploymentName,
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(1),
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
										Image: "daocloud.io/daocloud/dao-2048:latest",
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
			ginkgo.It(strconv.Itoa(testNumbers)+"th PVC STATUS should be Bound", func() {

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
			ginkgo.It(strconv.Itoa(testNumbers)+"th deploy STATUS should be AVAILABLE", func() {
				deployment := &appsv1.Deployment{}
				deployKey := k8sclient.ObjectKey{
					Name:      HaDeploymentName,
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
			ginkgo.It(strconv.Itoa(testNumbers)+"th Delete test Deployment", func() {
				//delete deploy
				deployment := &appsv1.Deployment{}
				deployKey := k8sclient.ObjectKey{
					Name:      HaDeploymentName,
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
			ginkgo.It(strconv.Itoa(testNumbers)+"th delete all pvc", func() {
				err := deleteAllPVC(ctx)
				gomega.Expect(err).To(gomega.BeNil())
			})
			ginkgo.It(strconv.Itoa(testNumbers)+"th check pv", func() {
				logrus.Printf("check pv")
				f := framework.NewDefaultFramework(ldapis.AddToScheme)
				client := f.GetClient()
				pvList := &apiv1.PersistentVolumeList{}

				err := wait.PollImmediate(3*time.Second, 3*time.Minute, func() (done bool, err error) {
					err = client.List(ctx, pvList)
					if err != nil {
						logrus.Error("get pv list error", err)
						f.ExpectNoError(err)
					}
					if len(pvList.Items) != 0 {
						return false, nil
					} else {
						return true, nil
					}
				})
				if err != nil {
					logrus.Error(err)
				}
				gomega.Expect(err).To(gomega.BeNil())
			})
		}
	})

	ginkgo.Context("Clean up the environment", func() {
		ginkgo.It("delete all sc", func() {
			err := deleteAllSC(ctx)
			gomega.Expect(err).To(gomega.BeNil())
		})
		ginkgo.It("delete helm", func() {
			uninstallHelm()

		})

	})

})
