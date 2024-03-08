package E2eTest

import (
	"context"
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"

	"github.com/hwameistor/hwameistor/test/e2e/framework"
	"github.com/hwameistor/hwameistor/test/e2e/utils"
	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = ginkgo.Describe("snapshot restore test ", ginkgo.Label("pvc"), func() {

	var f *framework.Framework
	var client ctrlclient.Client
	ctx := context.TODO()
	ginkgo.It("Configure the base environment", ginkgo.FlakeAttempts(5), func() {
		result := utils.ConfigureEnvironment(ctx)
		gomega.Expect(result).To(gomega.BeNil())
		f = framework.NewDefaultFramework(snapshotv1.AddToScheme)
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
			gomega.Expect(err).To(gomega.BeNil())
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
	ginkgo.Context("write test data", func() {
		ginkgo.It("add test file", func() {
			//create a request
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
	})
	ginkgo.Context("create a VolumeSnapshotClass", func() {
		ginkgo.It("create a VolumeSnapshotClass", func() {
			examplesnapshotclass := &snapshotv1.VolumeSnapshotClass{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeSnapshotClass",
					APIVersion: "snapshot.storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hwameistor-storage-lvm-snapshot",
					Namespace: "default",
					Annotations: map[string]string{
						"snapshot.storage.kubernetes.io/is-default-class": "true",
					},
				},
				Driver: "lvm.hwameistor.io",
				Parameters: map[string]string{
					"snapsize": "1073741824",
				},
				DeletionPolicy: "Delete",
			}
			err := client.Create(ctx, examplesnapshotclass)
			if err != nil {
				logrus.Printf("Create snapshotclass failed ：%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
	ginkgo.Context("create a VolumeSnapshot", func() {
		ginkgo.It("create a VolumeSnapshot", func() {
			PersistentVolumeClaimName := "pvc-lvm"
			VolumeSnapshotClassName := "hwameistor-storage-lvm-snapshot"
			examplesnapshot := &snapshotv1.VolumeSnapshot{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VolumeSnapshot",
					APIVersion: "snapshot.storage.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "snapshot-local-storage-pvc-lvm",
					Namespace: "default",
				},
				Spec: snapshotv1.VolumeSnapshotSpec{
					Source: snapshotv1.VolumeSnapshotSource{
						PersistentVolumeClaimName: &PersistentVolumeClaimName,
					},
					VolumeSnapshotClassName: &VolumeSnapshotClassName,
				},
				Status: nil,
			}
			err := client.Create(ctx, examplesnapshot)
			if err != nil {
				logrus.Printf("Create snapshot failed ：%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
	ginkgo.Context("check the Snapshot", func() {
		ginkgo.It("check the VolumeSnapshot", func() {
			logrus.Infof("Waiting for the VolumeSnapshot to be ready(2 minutes)")
			time.Sleep(2 * time.Minute)
			snapshot := &snapshotv1.VolumeSnapshot{}
			snapshotKey := ctrlclient.ObjectKey{
				Name:      "snapshot-local-storage-pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, snapshotKey, snapshot)
			if err != nil {
				logrus.Printf("get snapshot error :%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(snapshot.Status.ReadyToUse).To(gomega.Equal(utils.BoolPter(true)))
		})
		ginkgo.It("check the LocalVolumeSnapshot", func() {
			snapshot := &snapshotv1.VolumeSnapshot{}
			snapshotKey := ctrlclient.ObjectKey{
				Name:      "snapshot-local-storage-pvc-lvm",
				Namespace: "default",
			}
			err := client.Get(ctx, snapshotKey, snapshot)
			if err != nil {
				logrus.Printf("get snapshot error :%+v ", err)
				f.ExpectNoError(err)
			}
			localsnapshot := &v1alpha1.LocalVolumeSnapshot{}
			localsnapshotKey := ctrlclient.ObjectKey{
				Name:      *snapshot.Status.BoundVolumeSnapshotContentName,
				Namespace: "default",
			}
			f = framework.NewDefaultFramework(v1alpha1.AddToScheme)
			client = f.GetClient()
			err = client.Get(ctx, localsnapshotKey, localsnapshot)
			if err != nil {
				logrus.Printf("get snapshot error :%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(err).To(gomega.BeNil())
			logrus.Infof("check the state of VolumeSnapshot")
			err = wait.PollImmediate(3*time.Second, 1*time.Minute, func() (done bool, err error) {
				if err = client.Get(ctx, localsnapshotKey, localsnapshot); localsnapshot.Status.State != "Ready" {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				logrus.Printf("check the state of VolumeSnapshot error :%+v ", err)
				f.ExpectNoError(err)
			}
			gomega.Expect(localsnapshot.Status.State).To(gomega.Equal(v1alpha1.NodeStateReady))

		})
	})
	ginkgo.Context("create the restore PVC", func() {
		ginkgo.It("create the restore PVC", func() {
			//create PVC
			storageClassName := "local-storage-hdd-lvm"
			APIGroup := "snapshot.storage.k8s.io"
			restorePvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pvc-lvm-restore",
					Namespace: "default",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
					StorageClassName: &storageClassName,
					DataSource: &corev1.TypedLocalObjectReference{
						APIGroup: &APIGroup,
						Kind:     "VolumeSnapshot",
						Name:     "snapshot-local-storage-pvc-lvm",
					},
				},
			}
			err := client.Create(ctx, restorePvc)
			if err != nil {
				logrus.Printf("Create restorePvc failed ：%+v ", err)
				f.ExpectNoError(err)
			}

			gomega.Expect(err).To(gomega.BeNil())
		})

	})
	ginkgo.Context("create a deployment bound the restore PVC", func() {

		ginkgo.It("create deployment", func() {
			//create deployment
			exampleDeployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.DeploymentName + "-restore",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: utils.Int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "restore",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "restore",
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
											ClaimName: "pvc-lvm-restore",
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
				Name:      "pvc-lvm-restore",
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
				Name:      utils.DeploymentName + "-restore",
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
	ginkgo.Context("check the test data", func() {
		ginkgo.It("check test file", func() {
			//create a request
			config, err := config.GetConfig()
			if err != nil {
				return
			}

			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName + "-restore",
				Namespace: "default",
			}
			err = client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Printf("%+v ", err)
				f.ExpectNoError(err)
			}

			apps, err := labels.NewRequirement("app", selection.In, []string{"restore"})
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
		ginkgo.It("Delete restore Deployment", func() {
			//delete deploy
			deployment := &appsv1.Deployment{}
			deployKey := ctrlclient.ObjectKey{
				Name:      utils.DeploymentName + "-restore",
				Namespace: "default",
			}
			err := client.Get(ctx, deployKey, deployment)
			if err != nil {
				logrus.Error(err)
				f.ExpectNoError(err)
			}
			logrus.Infof("deleting restore Deployment ")

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
