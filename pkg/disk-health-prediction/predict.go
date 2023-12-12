package prediction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/smart/storage"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	HwameiNameSpace     = "hwameistor"
	SmartCMName         = "hwameistor-smart-result"
	PredictDataPrefix   = "predict-data-"
	PredictTMPPrefix    = "predict-tmp-result-"
	PredictResultPrefix = "predict-result-"
	DIR                 = "/data"
	NUM                 = 10
	DefaultNameSpace    = "default"
	ServiceAccountName  = "hwameistor-admin"
	ModelHostPath       = "/root/disk_model/model"
	ModelVolumePath     = "/disk_model"
	JobModelMountPath   = "/app/model/"
	JobVolumeName       = "model-volume"
	NodeFile            = "node-name"
	JobName             = "job-predict-"
)

type predictor struct {
	syncPeriod time.Duration
}

func NewPredictor() *predictor {
	return &predictor{}
}

func (p *predictor) WithSyncPeriod(syncPeriod time.Duration) *predictor {
	p.syncPeriod = syncPeriod
	return p
}

func (p *predictor) StartTimerPredict(ctx context.Context) {
	log.WithField("syncPeriod", p.syncPeriod).Info("Start disk timer predict")
	// timer trigger
	go wait.Until(p.predictAndSaveConfigMap, p.syncPeriod, ctx.Done())

	<-ctx.Done()
}

func (p *predictor) predictAndSaveConfigMap() {
	var (
		containerImage = "ghcr.io/hwameistor/ai-diskhealth-predict:v0.0.1"
		entryCommand   = "python predict_cm.py"
		err            error
	)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Error("Failed to create config object ")
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Error("Failed to create  Clientset")
		return
	}

	//1、Get disk smart-data from configmap(hwameistor-smart-result)
	dataFromSmart, err := collectionDataFromCm(clientset, SmartCMName, HwameiNameSpace)
	if err != nil {
		log.WithError(err).Error("Failed to Get disk smart-data from cm (hwameistor-smart-result)")
		return
	}
	//2、Extract training data set from smart data
	trainingData := extractTrainingdata(dataFromSmart)

	//3、Generate csv file
	err = createCSVFile(trainingData)
	if err != nil {
		log.WithError(err).Error("Failed to generate csv file")
		return
	}

	//4、create predict job
	err = createPredictJobs(DIR, clientset, containerImage, entryCommand)
	if err != nil {
		log.WithError(err).Error("Failed to create job to predict disk")
		return
	}
	//5、collect predict results
	err = collectPredictResults(clientset)
	if err != nil {
		log.WithError(err).Error("Fail to collect predict result")
		return
	}
}

func collectionDataFromCm(clientset *kubernetes.Clientset, cmName, ns string) (map[string][]Result, error) {

	cm, err := clientset.CoreV1().ConfigMaps(ns).Get(context.TODO(), cmName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Warnf("cm hwameistor-smart-result not found in default namespace\n")
			return nil, err
		} else {
			return nil, err
		}
	}

	data := make(map[string]string)
	if cm.Data == nil {
		return nil, fmt.Errorf("smart data is nil")
	}
	data = cm.Data
	totalResultMap := make(map[string][]Result)

	for node, value := range data {
		var totalResult []Result
		err := json.Unmarshal([]byte(value), &totalResult)
		if err != nil {
			return nil, err
		}
		totalResultMap[node] = totalResult
	}

	return totalResultMap, nil
}

func extractTrainingdata(totalResultMap map[string][]Result) map[string][]*DiskTrainingData {
	nodesDD := make(map[string][]*DiskTrainingData)
	for node, totalResult := range totalResultMap {
		var DD []*DiskTrainingData
		for _, result := range totalResult {
			tableMap := make(map[string]TableItem)
			for _, v := range result.AtaSmartAttributes.Table {
				tableMap[v.Name] = v
			}

			var disk DiskTrainingData
			if &result.Device != nil {
				disk.Node = node
				disk.Disk = result.Device.Name
				_, ok := tableMap[Smart_5_raw]
				if ok {
					disk.Smart_5_raw = tableMap[Smart_5_raw].Raw.Value
				}
				_, ok = tableMap[Smart_9_raw]
				if ok {
					disk.Smart_9_raw = tableMap[Smart_9_raw].Raw.Value
				}
				_, ok = tableMap[Smart_187_raw]
				if ok {
					disk.Smart_187_raw = tableMap[Smart_187_raw].Raw.Value
				}

				_, ok = tableMap[Smart_188_raw]
				if ok {
					disk.Smart_188_raw = tableMap[Smart_188_raw].Raw.Value
				}

				_, ok = tableMap[Smart_193_raw]
				if ok {
					disk.Smart_193_raw = tableMap[Smart_193_raw].Raw.Value
				}

				_, ok = tableMap[Smart_194_raw]
				if ok {
					disk.Smart_194_raw = tableMap[Smart_194_raw].Raw.Value
				}

				_, ok = tableMap[Smart_197_raw]
				if ok {
					disk.Smart_197_raw = tableMap[Smart_197_raw].Raw.Value
				}

				_, ok = tableMap[Smart_198_raw]
				if ok {
					disk.Smart_198_raw = tableMap[Smart_198_raw].Raw.Value
				}

				_, ok = tableMap[Smart_241_raw]
				if ok {
					disk.Smart_241_raw = tableMap[Smart_241_raw].Raw.Value
				}

				_, ok = tableMap[Smart_242_raw]
				if ok {
					disk.Smart_242_raw = tableMap[Smart_242_raw].Raw.Value
				}
				DD = append(DD, &disk)
			}
		}
		nodesDD[node] = DD
	}
	return nodesDD
}

func createCSVFile(data map[string][]*DiskTrainingData) error {

	err := os.MkdirAll(DIR, os.ModePerm)
	if err != nil {
		log.WithError(err).Error("Failed to mkdir /data  ")
		return err
	}
	for node, trainingData := range data {
		filename := node + ".csv"
		file, err := os.OpenFile(filepath.Join(DIR, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.WithError(err).Error("Failed create csv file  ")
			return err
		}

		if err = gocsv.MarshalFile(&trainingData, file); err != nil {
			log.WithError(err).Error("Failed write date into csv  ")
			return err
		}
		file.Close()
	}
	return nil
}

func createPredictJobs(dir string, clientset *kubernetes.Clientset, image, cmd string) error {
	var count int

	separator := string(os.PathSeparator)
	if _, err := os.Stat(DIR); !os.IsNotExist(err) {
		files, _ := os.ReadDir(DIR)
		for i, fileInfo := range files {
			count++
			filePath := dir + separator + fileInfo.Name()
			content, err := readFileContent(filePath)
			if err != nil {
				log.WithError(err).Error("Failed to read csv .")
				return err
			}
			smartStorage := storage.NewConfigMap(PredictDataPrefix+strings.Split(fileInfo.Name(), ".")[0], DefaultNameSpace).SetKubeClient(clientset)
			err = smartStorage.SetKV("CSV", string(content))
			if err != nil {
				log.WithError(err).Error("Failed to create predict cm .")
				return err
			}
			if i == len(files)-1 || count == NUM {
				//create a job
				jobName := JobName + time.Now().Format("2006010215")
				jobs := clientset.BatchV1().Jobs(HwameiNameSpace)
				var backOffLimit int32 = 0
				jobSpec := &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      jobName,
						Namespace: HwameiNameSpace,
					},
					Spec: batchv1.JobSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								ServiceAccountName: ServiceAccountName,
								Containers: []v1.Container{
									{
										Name:    jobName,
										Image:   image,
										Command: strings.Split(cmd, " "),
									},
								},
								RestartPolicy: v1.RestartPolicyNever,
							},
						},
						BackoffLimit: &backOffLimit,
					},
				}

				//If the user has trained the model, please configure disk-health-prediction pod host和volume-mount: /disk_model
				if _, err := os.Stat(ModelVolumePath); err == nil {
					content, err := os.ReadFile(ModelVolumePath + "/" + NodeFile)
					if err != nil {
						log.WithError(err).Error("Configure the hostname file as required")
					} else {
						nodeName := string(content)
						nodeName = strings.ReplaceAll(nodeName, "\n", "")
						jobSpec.Spec.Template.Spec.NodeName = nodeName
						if _, err := os.Stat(ModelVolumePath + "/model"); err == nil {
							var volumeMounts []v1.VolumeMount
							var v v1.VolumeMount
							v.MountPath = JobModelMountPath
							v.Name = JobVolumeName
							volumeMounts = append(volumeMounts, v)

							var volumes []v1.Volume
							var vol v1.Volume
							vol.Name = JobVolumeName
							volumeSource := v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: ModelHostPath,
								},
							}
							vol.VolumeSource = volumeSource
							volumes = append(volumes, vol)
							jobSpec.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
							jobSpec.Spec.Template.Spec.Volumes = volumes
						}
					}
				}

				_, err = jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
				if err != nil {
					log.WithError(err).Error("Failed to create K8s job.")
					return err
				}
				log.Infof("Created K8s job successfully")
				for {
					time.Sleep(time.Second * 5)
					job, err := jobs.Get(context.TODO(), jobName, metav1.GetOptions{})
					if err != nil {
						log.WithError(err).Error("get job error")
						return err
					}
					if job.Status.CompletionTime != nil {
						propagationPolicy := metav1.DeletePropagationBackground
						err = jobs.Delete(context.TODO(), jobName, metav1.DeleteOptions{PropagationPolicy: &propagationPolicy})
						if err != nil {
							log.WithError(err).Error("Failed to delete job that has finished .")
							return err
						}
						break
					}
				}
				// delete predict-data cm
				count = 0
				configMapList, err := clientset.CoreV1().ConfigMaps(DefaultNameSpace).List(context.TODO(), metav1.ListOptions{})
				if err != nil {
					log.WithError(err).Error("Failed to list predict-data cm .")
					return err
				}
				for _, cm := range configMapList.Items {
					if strings.Contains(cm.Name, PredictDataPrefix) {
						err = clientset.CoreV1().ConfigMaps(DefaultNameSpace).Delete(context.TODO(), cm.Name, metav1.DeleteOptions{})
						if err != nil {
							log.WithError(err).Error("Failed to delete predict-data cm .")
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func readFileContent(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func collectPredictResults(clientset *kubernetes.Clientset) error {
	results := make(map[string][]*DiskPredictResult)
	configMapList, err := clientset.CoreV1().ConfigMaps(DefaultNameSpace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed list config-map")
		return err
	}
	for _, cm := range configMapList.Items {
		if strings.Contains(cm.Name, PredictTMPPrefix) {
			data := cm.Data
			result := data["result"]
			var diskPredictResult []*DiskPredictResult
			err = gocsv.UnmarshalString(result, &diskPredictResult)
			if err != nil {
				log.WithError(err).Error("Failed unmarshal csv-data to struct-data")
				return err
			}

			results[cm.Name] = diskPredictResult
		}
	}

	for s, predictResults := range results {
		name := PredictResultPrefix + s[30:]
		smartStorage := storage.NewConfigMap(name, DefaultNameSpace).SetKubeClient(clientset)
		data, err := json.Marshal(&predictResults)
		if err != nil {
			log.WithError(err).Error("Failed json unmarshal")
			return err
		}
		now := time.Now().Format("2006-01-02_15.06.05")
		err = smartStorage.SetKV(now, string(data))
		if err != nil {
			log.WithError(err).Error("Failed to update predict result")
			return err
		}
	}

	for _, cm := range configMapList.Items {
		if strings.Contains(cm.Name, PredictTMPPrefix) {
			err = clientset.CoreV1().ConfigMaps(DefaultNameSpace).Delete(context.TODO(), cm.Name, metav1.DeleteOptions{})
			if err != nil {
				log.WithError(err).Error("Failed to delete tmp cm")
				return err
			}
		}
	}

	return nil
}
