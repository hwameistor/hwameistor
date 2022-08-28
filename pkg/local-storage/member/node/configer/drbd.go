package configer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apisv1alpha1 "github.com/hwameistor/hwameistor/pkg/apis/hwameistor/local-storage/v1alpha1"
	"github.com/hwameistor/hwameistor/pkg/local-storage/exechelper"
	"github.com/hwameistor/hwameistor/pkg/local-storage/exechelper/nsexecutor"
)

const (
	configDir      = "/etc/drbd.d"
	baseConfigPath = "/etc/drbd.conf"

	drbdDevicePrefix = "/dev/drbd"

	drbdadmCmd   = "drbdadm"
	drbdsetupCmd = "drbdsetup"
	drbdmetaCmd  = "drbdmeta"

	// disk state, https://www.linbit.com/drbd-user-guide/drbd-guide-9_0-en/#s-disk-states
	DiskStateUpToDate     = "UpToDate"
	DiskStateOutdated     = "Outdated"
	DiskStateInconsistent = "Inconsistent"
	DiskStateConsistent   = "Consistent"
	DiskStateDiskless     = "Diskless"
	DiskStateNegotiating  = "Negotiating"
	DiskStateDetaching    = "Detaching"
	DiskStateAttaching    = "Attaching"

	ConnectionStateConnected     = "Connected"
	ConnectionStateConnecting    = "Connecting"
	ConnectionStateDisconnecting = "Disconnecting"

	// replication state, https://www.linbit.com/drbd-user-guide/drbd-guide-9_0-en/#s-replication-states
	ReplicationEstablished = "Established"
	ReplicationSyncSource  = "SyncSource"
	ReplicationSyncTarget  = "SyncTarget"

	drbdMaxPeerCount = 3
)

var (
	configTmpl = `resource {{ .ResourceName }} {

  net {
    protocol C;
  }
{{ range .Peers }}

  on {{ .Hostname }} { 
    device    minor {{ $.Minor }};
    disk      {{ $.DevicePath }};
    address   {{ .IP }}:{{ $.Port }};
    meta-disk internal;
    node-id {{ .ID }};
  }
{{ end }}

  connection-mesh {
    hosts {{ range .Peers }} {{ .Hostname }} {{ end }};
  }
}`
)

type drbdConfig struct {
	ResourceName string
	Port         int
	Minor        int
	DevicePath   string
	Peers        []apisv1alpha1.VolumeReplica
}

type Resource struct {
	Name   string // Name = replica.Name
	Role   string
	Device struct {
		State string
	}
	Replication string
	PeerDevices map[string]*PeerDevice
}

type PeerDevice struct {
	NodeID int
	// ConnectionName=hostname
	ConnectionName string
	// local node state on the connection to this peer
	Replication string
	DiskState   string
}

type drbdConfigure struct {
	hostname       string
	apiClient      client.Client
	systemConfig   apisv1alpha1.SystemConfig
	statusSyncFunc SyncReplicaStatus
	cmdExec        exechelper.Executor
	lock           sync.Mutex
	once           sync.Once
	// record already applied configs on current node, key=replica.Name
	localConfigs map[string]apisv1alpha1.VolumeConfig
	// key=resource.Name
	resourceCache map[string]*Resource
	// resource.Name: replica.Name
	resourceReplicaNameMap map[string]string
	template               *template.Template
	logger                 *log.Entry
	stopCh                 <-chan struct{}
}

var _ Configer = &drbdConfigure{}

func NewDRBDConfiger(hostname string, systemConfig apisv1alpha1.SystemConfig, apiClient client.Client, syncFunc SyncReplicaStatus) (*drbdConfigure, error) {
	t, err := template.New("drbdConfigure").Parse(configTmpl)
	if err != nil {
		return nil, fmt.Errorf("parse drbd config template err: %s", err)
	}
	return &drbdConfigure{
		hostname:               hostname,
		apiClient:              apiClient,
		systemConfig:           systemConfig,
		cmdExec:                nsexecutor.New(),
		localConfigs:           make(map[string]apisv1alpha1.VolumeConfig),
		resourceCache:          make(map[string]*Resource),
		resourceReplicaNameMap: make(map[string]string),
		statusSyncFunc:         syncFunc,
		template:               t,
		logger:                 log.WithField("Module", "DRBDConfiger"),
	}, nil
}

func (m *drbdConfigure) Run(stopCh <-chan struct{}) {
	m.initConfigDirectory()

	m.stopCh = stopCh
}

func (m *drbdConfigure) initConfigDirectory() {
	if _, err := os.Stat(configDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(configDir, 0755); err != nil {
			m.logger.Fatalf("init config dir err: %s", err)
		}
	}
}

func (m *drbdConfigure) HasConfig(replica *apisv1alpha1.LocalVolumeReplica) bool {
	_, exists := m.resourceCache[m.genResourceName(replica)]
	return exists
}

func (m *drbdConfigure) IsConfigUpdated(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) bool {
	oldConfig, exists := m.localConfigs[replica.Name]
	if !exists {
		return true
	}

	return oldConfig.DeepEqual(&config)
}

func (m *drbdConfigure) ApplyConfig(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) error {
	m.logger.WithField("Replica", replica.Name).Infof("apply replica config")

	// start monitor when needed
	m.EnsureDRBDResourceStateMonitorStated()

	conf := m.config2DRBDConfig(replica, config)

	// create config file
	err := m.writeConfigFile(conf.ResourceName, conf)
	if err != nil {
		return fmt.Errorf("write config file err: %s", err)
	}

	// check if disk attached, only check/create metadata when disk unattached
	if dstate, err := m.getResourceDiskState(conf.ResourceName); err != nil || dstate == DiskStateDiskless {
		//m.logger.Debugf("get %s disk state: %s, err: %s", dstate, err)
		fmt.Printf("get %s disk state: %s, err: %s", conf.ResourceName, dstate, err)

		// check metadata
		if !m.hasMetadata(conf.Minor, conf.DevicePath) {
			// create metadata
			if err = m.createMetadata(conf.ResourceName, drbdMaxPeerCount); err != nil {
				return fmt.Errorf("create replica metadata err: %s", err)
			}
		}
	}

	// adjust for created or updated replica
	if err = m.adjustResource(conf.ResourceName); err != nil {
		return fmt.Errorf("adjust replica err: %s", err)
	}

	// create symbolic
	devicePath := m.getResourceDevicePath(conf)
	if _, err := os.Stat(replica.Status.DevicePath); err != nil && os.IsExist(err) {
		if err = os.Symlink(devicePath, replica.Status.DevicePath); err != nil {
			return fmt.Errorf("create symbolic link for %s err: %s", replica.Name, err)
		}
	}

	// resize resource
	if err = m.resizeResource(conf.ResourceName); err != nil {
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	m.localConfigs[replica.Name] = config
	m.resourceReplicaNameMap[conf.ResourceName] = replica.Name

	return nil
}

// Initialize do the initalization for volume
func (m *drbdConfigure) Initialize(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) error {
	m.logger.WithField("Replica", replica.Name).Info("initialize volume")
	resourceName := m.genResourceName(replica)
	if yes, err := m.isDeviceUpToDate(resourceName); err != nil {
		return fmt.Errorf("check resource dstate err: %s", err)
	} else if yes {
		m.logger.WithField("Replica", replica.Name).Infof("device is already UpToDate, skip initialize")
		return nil
	}

	if connected, err := m.isAllResourcePeersConnected(resourceName); err != nil {
		return err
	} else if !connected {
		return fmt.Errorf("not all resource peers are connected, can't do initialize")
	}

	return m.setResourceUpToDate(resourceName)
}

// ConsistencyCheck clean non-dlocal managed configs
func (m *drbdConfigure) ConsistencyCheck(replicas []apisv1alpha1.LocalVolumeReplica) {
	m.logger.Debug("do replica config ConsistencyCheck")
	knownResources := make(map[string]bool)
	for _, replica := range replicas {
		resoruceName := m.genResourceName(&replica)
		knownResources[resoruceName] = true
	}

	m.logger.Debugf("knownResources: %v", knownResources)

	// list node resource configs
	resourceNames, err := listNodeResourceNames()
	if err != nil {
		m.logger.Errorf("")
		return
	}

	// clean non-dlocal managed config
	for _, name := range resourceNames {
		if !knownResources[name] {
			m.logger.Infof("remove non-dlocal managed resource config: %s", name)
			if err = removeResourceConfigFile(name); err != nil {
				m.logger.Errorf("remove non-dlocal config file err: %s", err)
			}
		}
	}
}

func (m *drbdConfigure) writeConfigFile(resourceName string, conf drbdConfig) error {
	configPath := genConfigPath(resourceName)
	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open config file %s err: %s", configPath, err)
	}
	defer f.Close()

	if err = m.template.Execute(f, conf); err != nil {
		return fmt.Errorf("render and save config err: %s", err)
	}

	return nil
}

func genConfigPath(resourceName string) string {
	return path.Join(configDir, fmt.Sprintf("%s.res", resourceName))
}

// parseConfigFileName parse configFileName, return resourceName
func parseConfigFileName(configName string) string {
	return strings.TrimRight(configName, ".res")
}

// listNodeResourceNames list current node resources names by read config dir
func listNodeResourceNames() ([]string, error) {
	dir, err := os.Open(configDir)
	if err != nil {
		return nil, err
	}

	entries, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	var resourceNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if strings.HasSuffix(filename, ".res") {
			resourceNames = append(resourceNames, parseConfigFileName(filename))
		}
	}

	return resourceNames, nil
}

// removeResourceConfigFile return nil when configFile not exists
func removeResourceConfigFile(resourceName string) error {
	path := genConfigPath(resourceName)
	log.Debugf("remove path: %s", path)
	return os.RemoveAll(path)
}

func (m *drbdConfigure) createMetadata(resourceName string, peersCount int) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"create-md", resourceName, "--max-peers", strconv.Itoa(peersCount), "--force"},
		Timeout: 0,
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("create metadata for %s err: %d %s", resourceName, result.ExitCode, result.ErrBuf.String())
	}
	return nil
}

func (m *drbdConfigure) adjustResource(resourceName string) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"adjust", resourceName},
		Timeout: 0,
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("adjust %s err: %d, %s", resourceName, result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

func (m *drbdConfigure) getResourceDiskState(resourceName string) (string, error) {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"dstate", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return "", fmt.Errorf("get resource %s disk state err: %d, %s", resourceName, result.ExitCode, result.ErrBuf.String())
	}

	// there may be peers disk state
	return strings.SplitN(result.OutBuf.String(), "/", 2)[0], nil
}

// Show the current configuration of the resource
// if resource existed, config is not empty
func (m *drbdConfigure) showResource(resourceName string) (string, error) {
	cmd := exec.Command("nsenter", "-t", "1", "-n", "-u", "-i", "-m", "--",
		drbdsetupCmd, "show", resourceName)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	m.logger.Debugf("exec: %s %s", cmd.Path, cmd.Args)
	defer func() {
		m.logger.Debugf("stdout: %s stderr: %s", stdout.String(), stderr.String())
	}()

	var err error
	if err = cmd.Start(); err != nil {
		return "", err
	}
	if err = cmd.Wait(); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// resizeResource is idempotent
func (m *drbdConfigure) resizeResource(resourceName string) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"resize", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("resize resource %s err: %d, %s", resourceName, result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

func (m *drbdConfigure) getResourceDevicePath(conf drbdConfig) string {
	return fmt.Sprintf("%s%d", drbdDevicePrefix, conf.Minor)
}

func (m *drbdConfigure) DeleteConfig(replica *apisv1alpha1.LocalVolumeReplica) error {
	m.logger.WithField("Replica", replica.Name).Info("Delete Config")

	resourceName := m.genResourceName(replica)
	config, err := m.showResource(resourceName)
	if err != nil {
		return fmt.Errorf("show resource %s failed: %s", resourceName, err)
	}

	// only do down and wipe-md operation when resource is existed
	if config != "" {
		err := m.downResource(resourceName)
		if err != nil {
			return fmt.Errorf("down replica %s resource failed: %s", replica.Name, err)
		}

		if err = m.wipeMetadata(resourceName); err != nil {
			return fmt.Errorf("wipe replica resource metadata failed: %s", err)
		}
	}

	// remove symblink
	if err = os.Remove(replica.Status.DevicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove symbolic path %s err: %s", replica.Status.DevicePath, err)
	}

	// remove config file
	if err = removeResourceConfigFile(resourceName); err != nil {
		return fmt.Errorf("remove replica %s config file err: %s", replica.Name, err)
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.localConfigs, replica.Name)
	delete(m.resourceCache, resourceName)
	delete(m.resourceReplicaNameMap, resourceName)

	return nil
}

// call down on a resource multi times is safe
func (m *drbdConfigure) downResource(resourceName string) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"down", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("exit %d, %s", result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

// run wipeMetadata multi time for same resource is OK
func (m *drbdConfigure) wipeMetadata(resourceName string) error {
	cmd := exec.Command("nsenter", "-t", "1", "-n", "-u", "-i", "-m", "--",
		drbdadmCmd, "wipe-md", resourceName, "--force")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	m.logger.Debugf("exec: %s %s", cmd.Path, cmd.Args)
	defer func() {
		m.logger.Debugf("stdout: %s stderr: %s", stdout.String(), stderr.String())
	}()

	var err error
	if err = cmd.Start(); err != nil {
		return err
	}
	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (m *drbdConfigure) EnsureDRBDResourceStateMonitorStated() {
	m.once.Do(func() {
		go m.MonitorDRBDResourceState(m.stopCh)
	})
}

// monitor DRBD resource state and update replica status
func (m *drbdConfigure) MonitorDRBDResourceState(stopCh <-chan struct{}) error {
	for {
		select {
		case <-stopCh:
			{
				return nil
			}
		default:
		}

		if err := m.monitorDRBDResourceState(stopCh); err != nil {
			m.logger.Errorf("monitor drbd resource state err: %s, will retry later", err)
		}

		// avoid cmd exit&re-run too frequently
		time.Sleep(1 * time.Second)
	}
}

func (m *drbdConfigure) monitorDRBDResourceState(stopCh <-chan struct{}) error {
	m.logger.Info("start to monitor drbd resources")

	cmd := exec.Command("nsenter", "-t", "1", "-n", "-u", "-i", "-m", "--",
		drbdsetupCmd, "events2", "all")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("monitor drbd resource err: %s", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start drbd resource monitor err: %s", err)
	}
	go cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		event := scanner.Text()
		event = strings.TrimSpace(event)
		m.handleDRBDEvent(event)

		select {
		case <-stopCh:
			{
				return nil
			}
		default:
		}
	}

	return scanner.Err()
}

func (m *drbdConfigure) handleDRBDEvent(event string) {
	m.logger.WithField("event", event).Debugf("handle event")

	parts := strings.Split(event, " ")
	if len(parts) < 3 {
		m.logger.Debugf("ignore %s event: %s", drbdsetupCmd, event)
		return
	}
	nameParts := strings.SplitN(parts[2], ":", 2)
	if len(nameParts) != 2 {
		m.logger.Warnf("invalid name part: %s", parts[2])
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	resourceName := nameParts[1]
	resource, ok := m.resourceCache[resourceName]
	if !ok {
		resource = &Resource{
			Name:        resourceName,
			PeerDevices: make(map[string]*PeerDevice),
		}
		m.resourceCache[resourceName] = resource
	}
	eventMap := make(map[string]string)
	for _, part := range parts[2:] {
		p := strings.SplitN(part, ":", 2)
		if len(p) != 2 {
			continue
		}
		eventMap[p[0]] = p[1]
	}
	switch parts[1] {
	case "device":
		{
			if state, ok := eventMap["disk"]; ok {
				resource.Device.State = state
			}
		}
	case "peer-device":
		{
			var peerDevice *PeerDevice
			if hostname, ok := eventMap["conn-name"]; ok {
				if peerDevice, ok = resource.PeerDevices[hostname]; !ok {
					peerDevice = &PeerDevice{ConnectionName: hostname}
					resource.PeerDevices[hostname] = peerDevice
				}
			} else {
				return
			}
			if nodeID, ok := eventMap["peer-node-id"]; ok {
				id, _ := strconv.Atoi(nodeID)
				peerDevice.NodeID = id
			}
			if replication, ok := eventMap["replication"]; ok {
				peerDevice.Replication = replication
			}
			if diskState, ok := eventMap["peer-disk"]; ok {
				peerDevice.DiskState = diskState
			}
		}
	case "resource":
		{
			if role, ok := eventMap["role"]; ok {
				resource.Role = role
			}
		}
	default:
		m.logger.Debugf("ignore event: %s", event)
		// ignore other events
		return
	}

	replicaName := m.getReplicaName(resourceName)
	// ignore update for non-managed drbd resource
	if len(replicaName) == 0 {
		return
	}

	m.statusSyncFunc(replicaName)
}

func (m *drbdConfigure) getReplicaHAState(resource *Resource) apisv1alpha1.HAState {
	state := apisv1alpha1.HAState{State: apisv1alpha1.HAVolumeReplicaStateDown}
	switch resource.Device.State {
	case DiskStateUpToDate:
		state.State = apisv1alpha1.HAVolumeReplicaStateConsistent
	case DiskStateInconsistent, DiskStateConsistent, DiskStateOutdated:
		state.State = apisv1alpha1.HAVolumeReplicaStateInconsistent
	case DiskStateNegotiating:
		state.State = apisv1alpha1.HAVolumeReplicaStateUp
	case DiskStateDiskless, DiskStateDetaching, DiskStateAttaching:
		state.State = apisv1alpha1.HAVolumeReplicaStateDown
	}

	state.Reason = fmt.Sprintf("device is %s", resource.Device.State)
	return state
}

func (m *drbdConfigure) GetReplicaHAState(replica *apisv1alpha1.LocalVolumeReplica) (apisv1alpha1.HAState, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	resourceName := m.genResourceName(replica)
	resource, ok := m.resourceCache[resourceName]
	if !ok {
		return apisv1alpha1.HAState{}, fmt.Errorf("replica %s not found in local cache", replica.Name)
	}
	haState := m.getReplicaHAState(resource)
	return haState, nil
}

func (m *drbdConfigure) hasMetadata(minor int, devicePath string) bool {
	// force is needed if the drbd-resource is still in Negotiating state or earlier.
	// in that case, drbdmeta asks "Exclusive open failed. Do it anyways?" and expects to type 'yes'.
	// should not break anything
	params := exechelper.ExecParams{
		CmdName: drbdmetaCmd,
		CmdArgs: []string{strconv.Itoa(minor), "v09", devicePath, "internal", "get-gi", "--node-id", "0", "--force"},
		Timeout: 0,
	}
	result := m.cmdExec.RunCommand(params)
	return result.ExitCode == 0
}

func (m *drbdConfigure) config2DRBDConfig(replica *apisv1alpha1.LocalVolumeReplica, config apisv1alpha1.VolumeConfig) drbdConfig {
	port := config.ResourceID + m.systemConfig.DRBD.StartPort
	return drbdConfig{
		ResourceName: m.genResourceName(replica),
		Port:         port,
		Minor:        port,
		DevicePath:   replica.Status.StoragePath,
		Peers:        config.Replicas,
	}
}

func (m *drbdConfigure) isDeviceUpToDate(resourceName string) (bool, error) {
	state, err := m.getDeviceState(resourceName)
	if err != nil {
		return false, fmt.Errorf("get device state err: %s", err)
	}

	return state == DiskStateUpToDate, nil
}

func (m *drbdConfigure) getDeviceState(resourceName string) (string, error) {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"dstate", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return "", result.Error
	}

	content := result.OutBuf.String()
	state := strings.SplitN(content, "/", 2)[0]
	return state, nil
}

func (m *drbdConfigure) setResourceUpToDate(resourceName string) error {
	// use new-current-uuid instead of primary --force to avoid a fully
	// device resynchronization, it's too slow. (2m+ for 10Gi resource)
	return m.newCurrentUUID(resourceName, true)
}

func (m *drbdConfigure) isAllResourcePeersConnected(resourceName string) (bool, error) {
	cstates, err := m.getResourceConnectionState(resourceName)
	if err != nil {
		return false, err
	}

	for _, cstate := range cstates {
		// connection state = "" if only one replica
		if len(cstate) != 0 && cstate != ConnectionStateConnected {
			return false, nil
		}
	}
	return true, nil
}

func (m *drbdConfigure) getResourceConnectionState(resourceName string) ([]string, error) {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"cstate", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("get resource cstate err: %d, %s", result.ExitCode, result.ErrBuf.String())
	}

	/*	empty string when resource is single peer, else output one peer state one line, eg:
		Connected
		Connecting
	*/
	return strings.Split(result.OutBuf.String(), "\n"), nil
}

func (m *drbdConfigure) newCurrentUUID(resourceName string, clearBitmap bool) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"new-current-uuid", resourceName},
	}
	if clearBitmap {
		params.CmdArgs = append(params.CmdArgs, "--clear-bitmap")
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("new-current-uuid resource err: %d, %s", result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

func (m *drbdConfigure) primaryResource(resourceName string, force bool) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"primary", resourceName},
	}
	if force {
		params.CmdArgs = append(params.CmdArgs, "--force")
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("primary resource err: %d, %s", result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

func (m *drbdConfigure) secondaryResource(resourceName string) error {
	params := exechelper.ExecParams{
		CmdName: drbdadmCmd,
		CmdArgs: []string{"secondary", resourceName},
	}
	result := m.cmdExec.RunCommand(params)
	if result.ExitCode != 0 {
		return fmt.Errorf("secondary resource err: %d, %s", result.ExitCode, result.ErrBuf.String())
	}

	return nil
}

func (m *drbdConfigure) genResourceName(replica *apisv1alpha1.LocalVolumeReplica) string {
	return replica.Spec.VolumeName
}

func (m *drbdConfigure) getReplicaName(resourceName string) string {
	return m.resourceReplicaNameMap[resourceName]
}

func (m *drbdConfigure) isPrimary(config apisv1alpha1.VolumeConfig) bool {
	for _, peer := range config.Replicas {
		if peer.Hostname == m.hostname && peer.Primary {
			return true
		}
	}

	return false
}
