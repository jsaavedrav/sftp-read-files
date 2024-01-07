package readfiles

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/pkg/sftp"
)

const (
	sftpUser = "sftp_user"
	sftpPass = "sftp1209"
	sftpHost = "127.0.0.1"
	sftpPort = "2036"
)

type remoteFiles struct {
	Name    string
	Size    string
	ModTime string
	Value   string
}

func ReadFiles() {
	// init sftp
	rawurl := fmt.Sprintf("sftp://%v:%v@%v", sftpUser, sftpPass, sftpHost)

    // Parse the URL 
    parsedUrl, err := url.Parse(rawurl)
    if err != nil {
        log.Fatalf("Failed to parse SFTP To Go URL: %s", err)
    }

    // Get user name and pass
    user := parsedUrl.User.Username()
    pass, _ := parsedUrl.User.Password()

    // Parse Host and Port
    host := parsedUrl.Host

    // Get hostkey 
    // hostKey := getHostKey(host)

    log.Printf("Connecting to %s ...\n", host)

    var auths []ssh.AuthMethod

    // Try to use $SSH_AUTH_SOCK which contains the path of the unix file socket that the sshd agent uses
    // for communication with other processes.
    if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
    }

    // Use password authentication if provided
    if pass != "" {
        auths = append(auths, ssh.Password(pass))
    }

    // Initialize client configuration
    config := ssh.ClientConfig{
        User: user,
        Auth: auths,
        // Auth: []ssh.AuthMethod{
        //  ssh.KeyboardInteractive(SshInteractive),
        // },

        // Uncomment to ignore host key check
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        // HostKeyCallback: ssh.FixedHostKey(hostKey),
        // HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
        //  return nil
        // },
        Timeout: 30 * time.Second,
    }

    addr := fmt.Sprintf("%s:%s", host, sftpPort)

    // Connect to server
    conn, err := ssh.Dial("tcp", addr, &config)
    if err != nil {
        log.Fatalf("Failed to connec to host [%s]: %v", addr, err)
    }

    defer conn.Close()

    // Create new SFTP client
    sc, err := sftp.NewClient(conn)
    if err != nil {
        log.Fatalf("Unable to start SFTP subsystem: %v", err)
    }
    defer sc.Close()

    // List files in the root directory .
    theFiles, err := listFiles(sc, "uploads/.")
    if err != nil {
        log.Fatalf("failed to list files in .: %v", err)
    }

    log.Printf("Found Files in . Files")
    // Output each file name and size in bytes
    log.Printf("%19s %12s %s", "MOD TIME", "SIZE", "NAME")
    for _, theFile := range theFiles {
        // log.Printf("%19s %12s %s", theFile.ModTime, theFile.Size, theFile.Name)
        txtFileCont, err := readFile(sc, "uploads/"+theFile.Name)
        if err != nil {
            log.Fatalf("Could not read file data.csv; %v", err)
        }
        fmt.Print(txtFileCont)
    }

    // txtFile, err := readFile(sc, "uploads/test.txt")
    // if err != nil {
    //     log.Fatalf("Could not read file data.csv; %v", err)
    // }
    // fmt.Print(txtFile)

    return 
}



// read file from sftp server
func readFile(sc* sftp.Client, remoteFile string) (txtFile remoteFiles, err error){
	// Note: SFTP To Go doesn't support O_RDWR mode
	srcFile, err := sc.OpenFile(remoteFile, (os.O_RDONLY))
	if err != nil {
			fmt.Printf("unable to open remote file: %v", err)
	}
	defer srcFile.Close()
	b, err := io.ReadAll(srcFile)
	txt := string(b)
	txtFile = remoteFiles{
			Name:    remoteFile,
			Size:    "",
			ModTime: "",
			Value:  txt,
	}
	fmt.Print(txtFile)
	return txtFile, nil
}

func listFiles(sc* sftp.Client, remoteDir string) (theFiles []remoteFiles, err error) {

	files, err := sc.ReadDir(remoteDir)
	if err != nil {
			return theFiles, fmt.Errorf("Unable to list remote dir: %v", err)
	}

	for _, f := range files {
			var name, modTime, size string

			name = f.Name()
			modTime = f.ModTime().Format("2006-01-02 15:04:05")
			size = fmt.Sprintf("%12d", f.Size())

			if f.IsDir() {
					name = name + "/"
					modTime = ""
					size = "PRE"
			}

			theFiles = append(theFiles, remoteFiles{
					Name:    name,
					Size:    size,
					ModTime: modTime,
			})
	}

	return theFiles, nil
}