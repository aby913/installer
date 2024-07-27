package patch

import (
	"fmt"
	"path"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/binaries"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

// ~ PatchTask apt-get install
type PatchTask struct {
	common.KubeAction
}

func (t *PatchTask) Execute(runtime connector.Runtime) error {
	var cmd string
	var debianFrontend string
	var pre_reqs = "apt-transport-https ca-certificates curl"

	if _, err := util.GetCommand(common.CommandGPG); err != nil {
		pre_reqs = pre_reqs + " gnupg"
	}

	switch constants.OsPlatform {
	case common.Debian:
		debianFrontend = "DEBIAN_FRONTEND=noninteractive"
		fallthrough
	case common.Ubuntu, common.Raspbian:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s update -qq", constants.PkgManager), false, false); err != nil {
			logger.Errorf("update os error %v", err)
			return err
		}

		if _, err := runtime.GetRunner().SudoCmd("apt --fix-broken install -y", false, false); err != nil {
			logger.Errorf("fix-broken install error %v", err)
			return err
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s %s install -y -qq %s", debianFrontend, constants.PkgManager, pre_reqs), false, false); err != nil {
			logger.Errorf("install deps %s error %v", pre_reqs, err)
			return err
		}

		var cmd = "conntrack socat apache2-utils ntpdate net-tools make gcc openssh-server bison flex"
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s %s install -y %s", debianFrontend, constants.PkgManager, cmd), false, false); err != nil {
			logger.Errorf("install deps %s error %v", cmd, err)
			return err
		}
	case common.CentOs, common.Fedora, common.RHEl:
		cmd = "conntrack socat httpd-tools ntpdate net-tools make gcc openssh-server"
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s install -y %s", constants.PkgManager, cmd), false, false); err != nil {
			logger.Errorf("install deps %s error %v", cmd, err)
			return err
		}
	}

	return nil
}

// ~ SocatTask
type SocatTask struct {
	common.KubeAction
}

func (t *SocatTask) Execute(runtime connector.Runtime) error {
	filePath, fileName, err := binaries.DownloadSocat(runtime.GetWorkDir(), kubekeyapiv1alpha2.DefaultSocatVersion, constants.OsArch, t.PipelineCache)
	if err != nil {
		logger.Errorf("failed to download socat: %v", err)
		return err
	}
	f := path.Join(filePath, fileName)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", f, filePath), false, false); err != nil {
		logger.Errorf("failed to extract %s %v", f, err)
		return err
	}

	tp := path.Join(filePath, fmt.Sprintf("socat-%s", kubekeyapiv1alpha2.DefaultSocatVersion))
	if err := util.ChangeDir(tp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("./configure --prefix=/usr && make -j4 && make install && strip socat", false, false); err != nil {
			logger.Errorf("failed to install socat %v", err)
			return err
		}
	}
	if err := util.ChangeDir(runtime.GetRootDir()); err != nil {
		logger.Errorf("failed to change dir %v", err)
		return err
	}

	return nil
}

// ~ ConntrackTask
type ConntrackTask struct {
	common.KubeAction
}

func (t *ConntrackTask) Execute(runtime connector.Runtime) error {
	flexFilePath, flexFileName, err := binaries.DownloadFlex(runtime.GetWorkDir(), kubekeyapiv1alpha2.DefaultFlexVersion, constants.OsArch, t.PipelineCache)
	if err != nil {
		logger.Errorf("failed to download flex: %v", err)
		return err
	}
	filePath, fileName, err := binaries.DownloadConntrack(runtime.GetWorkDir(), kubekeyapiv1alpha2.DefaultConntrackVersion, constants.OsArch, t.PipelineCache)
	if err != nil {
		logger.Errorf("failed to download conntrack: %v", err)
		return err
	}
	fl := path.Join(flexFilePath, flexFileName)
	f := path.Join(filePath, fileName)

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", fl, filePath), false, true); err != nil {
		logger.Errorf("failed to extract %s %v", flexFilePath, err)
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar xzvf %s -C %s", f, filePath), false, true); err != nil {
		logger.Errorf("failed to extract %s %v", f, err)
		return err
	}

	// install
	fp := path.Join(flexFilePath, fmt.Sprintf("flex-%s", kubekeyapiv1alpha2.DefaultFlexVersion))
	if err := util.ChangeDir(fp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("autoreconf -i && ./configure --prefix=/usr && make -j4 && make install", false, true); err != nil {
			logger.Errorf("failed to install flex %v", err)
			return err
		}
	}

	tp := path.Join(filePath, fmt.Sprintf("conntrack-tools-conntrack-tools-%s", kubekeyapiv1alpha2.DefaultConntrackVersion))
	if err := util.ChangeDir(tp); err == nil {
		if _, err := runtime.GetRunner().SudoCmd("autoreconf -i && ./configure --prefix=/usr && make -j4 && make install", false, true); err != nil {
			logger.Errorf("failed to install conntrack %v", err)
			return err
		}
	}
	if err := util.ChangeDir(runtime.GetRootDir()); err != nil {
		logger.Errorf("failed to change dir %v", err)
		return err
	}

	return nil
}