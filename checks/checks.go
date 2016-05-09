package checks

import (
	"strings"

	"github.com/docker/engine-api/client"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	//"log"
)

type Result struct {
	Name   string
	Status string
	Output string
}

//Skip is used when a check won't run. Output is used to describe the reason.
func (r *Result) Skip(s string) {
	r.Status = "SKIP"
	r.Output = s
	return
}

//Pass is used when a check has passed
func (r *Result) Pass() {
	r.Status = "PASS"
	return
}

//Fail is used when a check has failed. Output is used to describe the reason.
func (r *Result) Fail(s string) {
	r.Status = "WARN"
	r.Output = s
	return
}

func (r *Result) Info(s string) {
	r.Status = "INFO"
	r.Output = s
	return
}

type Check func(client *client.Client) Result

var checklist = map[string]Check{
	//Docker Host
	"kernel_version":     CheckKernelVersion,
	"separate_partition": CheckSeparatePartion,
	"running_services":   CheckRunningServices,
	"server_version":     CheckDockerVersion,
	"trusted_users":      CheckTrustedUsers,
	"audit_daemon":       AuditDockerDaemon,
	"audit_lib":          AuditLibDocker,
	"audit_etc":          AuditEtcDocker,
	"audit_registry":     AuditDockerRegistry,
	"audit_service":      AuditDockerService,
	"audit_socket":       AuditDockerSocket,
	"audit_sysconfig":    AuditDockerSysconfig,
	"audit_network":      AuditDockerNetwork,
	"audit_sysregistry":  AuditDockerSysRegistry,
	"audit_storage":      AuditDockerStorage,
	"audit_default":      AuditDockerDefault,
	//Docker Files
	"docker.service_perms":          CheckServicePerms,
	"docker.service_owner":          CheckServiceOwner,
	"docker-registry.service_owner": CheckRegistryOwner,
	"docker-registry.service_perms": CheckRegistryPerms,
	"docker.socket_owner":           CheckSocketOwner,
	"docker.socket_perms":           CheckSocketPerms,
	"dockerenv_owner":               CheckEnvOwner,
	"dockerenv_perms":               CheckEnvPerms,
	"docker-network_owner":          CheckNetEnvOwner,
	"docker-network_perms":          CheckNetEnvPerms,
	"docker-registry_owner":         CheckRegEnvOwner,
	"docker-registry_perms":         CheckRegEnvPerms,
	"docker-storage_owner":          CheckStoreEnvOwner,
	"docker-storage_perms":          CheckStoreEnvPerms,
	"dockerdir_owner":               CheckDockerDirOwner,
	"dockerdir_perms":               CheckDockerDirPerms,
	"registrycerts_owner":           CheckRegistryCertOwner,
	"registrycerts_perms":           CheckRegistryCertPerms,
	"cacert_owner":                  CheckCACertOwner,
	"cacert_perms":                  CheckCACertPerms,
	"servercert_owner":              CheckServerCertOwner,
	"servercert_perms":              CheckServerCertPerms,
	"certkey_owner":                 CheckCertKeyOwner,
	"certkey_perms":                 CheckCertKeyPerms,
	"socket_owner":                  CheckDockerSockOwner,
	"socket_perms":                  CheckDockerSockPerms,
	//Docker Configuration
	"lxc_driver":        CheckLxcDriver,
	"net_traffic":       RestrictNetTraffic,
	"logging_level":     CheckLoggingLevel,
	"allow_iptables":    CheckIpTables,
	"insecure_registry": CheckInsecureRegistry,
	"local_registry":    CheckLocalRegistry,
	"aufs_driver":       CheckAufsDriver,
	"default_socket":    CheckDefaultSocket,
	"tls_auth":          CheckTLSAuth,
	"default_ulimit":    CheckUlimit,
	//Docker Container Images
	"root_containers": CheckContainerUser,
	//Docker Container Runtime
	"apparmor_profile":      CheckAppArmor,
	"selinux_options":       CheckSELinux,
	"single_process":        CheckSingleMainProcess,
	"kernel_capabilities":   CheckKernelCapabilities,
	"privileged_containers": CheckPrivContainers,
	"sensitive_dirs":        CheckSensitiveDirs,
	"ssh_running":           CheckSSHRunning,
	"privileged_ports":      CheckPrivilegedPorts,
	"needed_ports":          CheckNeededPorts,
	"host_net_mode":         CheckHostNetworkMode,
	"memory_usage":          CheckMemoryLimits,
	"cpu_shares":            CheckCPUShares,
	"readonly_rootfs":       CheckReadonlyRoot,
	"bind_specific_int":     CheckBindHostInterface,
	"restart_policy":        CheckRestartPolicy,
	"host_namespace":        CheckHostNamespace,
	"ipc_namespace":         CheckIPCNamespace,
	"host_devices":          CheckHostDevices,
	"override_ulimit":       CheckDefaultUlimit,
	//Docker Security Operations
	"central_logging":  CheckCentralLogging,
	"container_sprawl": CheckContainerSprawl,
}

func GetAuditDefinitions() map[string]Check {

	return checklist
}

func GetProcCmdline(procname string) (cmd []string, err error) {
	var pid int

	ps, _ := ps.Processes()
	for i, _ := range ps {
		if ps[i].Executable() == procname {
			pid = ps[i].Pid()
			break
		}
	}
	proc, err := process.NewProcess(int32(pid))
	cmd, err = proc.CmdlineSlice()
	return cmd, err
}

func GetCmdOption(args []string, opt string) (exist bool, val string) {
	var optBuf string
	for _, arg := range args {
		if strings.Contains(arg, opt) {
			optBuf = arg
			exist = true
			break
		}
	}
	if exist {
		nameVal := strings.Split(optBuf, "=")
		if len(nameVal) > 1 {
			val = strings.TrimSuffix(nameVal[1], " ")
		}
	} else {
		exist = false
	}

	return exist, val
}