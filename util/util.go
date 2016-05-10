package util

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/akutz/gotil"

	"github.com/emccode/polly/core/version"
)

const (
	logDirPathSuffix = "/var/log/polly"
	etcDirPathSuffix = "/etc/polly"
	binDirPathSuffix = "/usr/bin"
	runDirPathSuffix = "/var/run/polly"
	libDirPathSuffix = "/var/lib/polly"

	// UnitFilePath is the path to the SystemD service's unit file.
	UnitFilePath = "/etc/systemd/system/polly.service"

	// InitFilePath is the path to the SystemV Service's init script.
	InitFilePath = "/etc/init.d/polly"

	// EnvFileName is the name of the environment file used by the SystemD
	// service.
	EnvFileName = "polly.env"
)

var (
	thisExeDir     string
	thisExeName    string
	thisExeAbsPath string

	prefix string

	binDirPath  string
	binFilePath string
	logDirPath  string
	libDirPath  string
	runDirPath  string
	etcDirPath  string
	pidFilePath string
)

func init() {
	prefix = os.Getenv("POLLY_HOME")

	thisExeDir, thisExeName, thisExeAbsPath = gotil.GetThisPathParts()
}

// GetPrefix gets the root path to the REX-Ray data.
func GetPrefix() string {
	return prefix
}

// Prefix sets the root path to the REX-Ray data.
func Prefix(p string) {
	if p == "" || p == "/" {
		return
	}

	binDirPath = ""
	binFilePath = ""
	logDirPath = ""
	libDirPath = ""
	runDirPath = ""
	etcDirPath = ""
	pidFilePath = ""

	prefix = p
}

// IsPrefixed returns a flag indicating whether or not a prefix value is set.
func IsPrefixed() bool {
	return !(prefix == "" || prefix == "/")
}

// Install executes the system install command.
func Install(args ...string) {
	exec.Command("install", args...).Run()
}

// InstallChownRoot executes the system install command and chowns the target
// to the root user and group.
func InstallChownRoot(args ...string) {
	a := []string{"-o", "0", "-g", "0"}
	for _, i := range args {
		a = append(a, i)
	}
	exec.Command("install", a...).Run()
}

// InstallDirChownRoot executes the system install command with a -d flag and
// chowns the target to the root user and group.
func InstallDirChownRoot(dirPath string) {
	InstallChownRoot("-d", dirPath)
}

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func readable(path string) bool {
	return unix.Access(path, unix.R_OK) == nil
}

// EtcDirPath returns the path to the REX-Ray etc directory.
func EtcDirPath() string {
	if etcDirPath == "" {
		etcDirPath = fmt.Sprintf("%s%s", prefix, etcDirPathSuffix)
		err := os.MkdirAll(etcDirPath, 0755)
		if err != nil || !readable(etcDirPath) {
			log.WithError(err).Warning("could not access etc dir")
		}

	}
	return etcDirPath
}

// RunDirPath returns the path to the REX-Ray run directory.
func RunDirPath() string {
	if runDirPath == "" {
		runDirPath = fmt.Sprintf("%s%s", prefix, runDirPathSuffix)
		err := os.MkdirAll(runDirPath, 0755)
		if err != nil || !readable(runDirPath) || !writable(runDirPath) {
			log.WithError(err).WithField("runDirPath", runDirPath).Warning("missing r|w access run dir")
		}

	}
	return runDirPath
}

// LogDirPath returns the path to the REX-Ray log directory.
func LogDirPath() string {
	if logDirPath == "" {
		logDirPath = fmt.Sprintf("%s%s", prefix, logDirPathSuffix)
		err := os.MkdirAll(logDirPath, 0755)
		if err != nil || !readable(logDirPath) || !writable(logDirPath) {
			log.WithError(err).WithField("logDirPath", logDirPath).Warning("missing r|w access log dir")
		}
	}
	return logDirPath
}

// LibDirPath returns the path to the REX-Ray bin directory.
func LibDirPath() string {
	if libDirPath == "" {
		libDirPath = fmt.Sprintf("%s%s", prefix, libDirPathSuffix)
		os.MkdirAll(libDirPath, 0755)
	}
	return libDirPath
}

// LibFilePath returns the path to a file inside the REX-Ray lib directory
// with the provided file name.
func LibFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LibDirPath(), fileName)
}

// BinDirPath returns the path to the REX-Ray bin directory.
func BinDirPath() string {
	if binDirPath == "" {
		binDirPath = fmt.Sprintf("%s%s", prefix, binDirPathSuffix)
		os.MkdirAll(binDirPath, 0755)
	}
	return binDirPath
}

// PidFilePath returns the path to the REX-Ray PID file.
func PidFilePath() string {
	if pidFilePath == "" {
		pidFilePath = fmt.Sprintf("%s/polly.pid", RunDirPath())
	}
	return pidFilePath
}

// BinFilePath returns the path to the REX-Ray executable.
func BinFilePath() string {
	if binFilePath == "" {
		binFilePath = fmt.Sprintf("%s/polly", BinDirPath())
	}
	return binFilePath
}

// EtcFilePath returns the path to a file inside the REX-Ray etc directory
// with the provided file name.
func EtcFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", EtcDirPath(), fileName)
}

// LogFilePath returns the path to a file inside the REX-Ray log directory
// with the provided file name.
func LogFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", LogDirPath(), fileName)
}

// LogFile returns a writer to a file inside the REX-Ray log directory
// with the provided file name.
func LogFile(fileName string) (io.Writer, error) {
	return os.OpenFile(
		LogFilePath(fileName), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
}

// StdOutAndLogFile returns a mutltiplexed writer for the current process's
// stdout descriptor and a REX-Ray log file with the provided name.
func StdOutAndLogFile(fileName string) (io.Writer, error) {
	lf, lfErr := LogFile(fileName)
	if lfErr != nil {
		return nil, lfErr
	}
	return io.MultiWriter(os.Stdout, lf), nil
}

// WritePidFile writes the current process ID to the REX-Ray PID file.
func WritePidFile(pid int) error {

	if pid < 0 {
		pid = os.Getpid()
	}

	return gotil.WriteStringToFile(fmt.Sprintf("%d", pid), PidFilePath())
}

// ReadPidFile reads the REX-Ray PID from the PID file.
func ReadPidFile() (int, error) {

	pidStr, pidStrErr := gotil.ReadFileToString(PidFilePath())
	if pidStrErr != nil {
		return -1, pidStrErr
	}

	pid, atoiErr := strconv.Atoi(pidStr)
	if atoiErr != nil {
		return -1, atoiErr
	}

	return pid, nil
}

// PrintVersion prints the current version information to the provided writer.
func PrintVersion(out io.Writer) {
	fmt.Fprintf(out, "Binary: %s\n", thisExeAbsPath)
	fmt.Fprintf(out, "SemVer: %s\n", version.SemVer)
	fmt.Fprintf(out, "OsArch: %s\n", version.Arch)
	fmt.Fprintf(out, "Branch: %s\n", version.Branch)
	fmt.Fprintf(out, "Commit: %s\n", version.ShaLong)
	fmt.Fprintf(out, "Formed: %s\n", version.EpochToRfc1123())
}

//ContainsString searches an array for a matching string
func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
