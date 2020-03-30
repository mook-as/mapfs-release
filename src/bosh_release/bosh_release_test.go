package bosh_release_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"time"
)

var _ = Describe("BoshReleaseTest", func() {
	BeforeEach(func() {
		deploy()
	})

	It("should have the mapfs binaries", func() {
		expectFileInstalled("/var/vcap/packages/mapfs/bin/mapfs")
		expectDpkgInstalled("ii  libfuse2:amd64.*2.9.7-1+deb9u2")
		expectDpkgInstalled("ii  fuse.*2.9.7-1+deb9u2")
		expectFileInstalled("/etc/fuse.conf")
	})

	Context("when upgrading from an older version of fuse", func() {
		BeforeEach(func() {
			undeploy()
			deploy("./operations/remove-mapfs-job.yaml")

			scp("./assets/fuse_2.9.4-1ubuntu3.1_amd64.deb", "/tmp/fuse.deb")
			scp("./assets/libfuse2_2.9.4-1ubuntu3.1_amd64.deb", "/tmp/libfuse.deb")
			installDpkg("/tmp/libfuse.deb")
			installDpkg("/tmp/fuse.deb")

			expectDpkgInstalled("ii  libfuse2:amd64.*2.9.4")
			expectDpkgInstalled("ii  fuse.*2.9.4")

			deploy()
		})

		It("should have the mapfs binaries", func() {
			expectFileInstalled("/var/vcap/packages/mapfs/bin/mapfs")
			expectDpkgInstalled("ii  libfuse2:amd64.*2.9.7-1+deb9u2")
			expectDpkgInstalled("ii  fuse.*2.9.7-1+deb9u2")
			expectFileInstalled("/etc/fuse.conf")
		})

	})

	Context("when mapfs is disabled", func() {

		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "delete-deployment", "-n")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			deploy("./operations/disable-mapfs.yml")
		})

		It("should not have or configured the fuse package", func() {
			expectFileInstalled("/var/vcap/packages/mapfs/bin/mapfs")
			expectDpkgNotInstalled(" fuse ")
			expectFileNotInstalled("/etc/fuse.conf")
		})
	})

	Context("when another process has a dpkg lock", func() {

		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo dpkg -P fuse")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo rm -f /tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "scp", dpkgLockBuildPackagePath, "mapfs:/tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo /tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("locked /var/lib/dpkg/lock"))
		})

		AfterEach(func() {
			releaseDpkgLock()
		})

		It("should successfully dpkg install", func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo /var/vcap/jobs/mapfs/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("dpkg: error: dpkg status database is locked by another process"))
			releaseDpkgLock()
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should eventually timeout when the dpkg lock is not released in a reasonable time", func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo /var/vcap/jobs/mapfs/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("dpkg: error: dpkg status database is locked by another process"))
			Eventually(session, 6*time.Minute, 1*time.Second).Should(gexec.Exit(1))
		})
	})

})

func releaseDpkgLock() {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo pkill lock_dpkg")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func expectDpkgNotInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("dpkg -l | grep '%s'", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), fmt.Sprintf("package [%s] was found when it should not have.", dpkgName))
}

func expectDpkgInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("dpkg -l | grep '%s'", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), fmt.Sprintf("package [%s] was not found when it should have.", dpkgName))
}

func expectFileInstalled(filePath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("stat %s", filePath))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), fmt.Sprintf("file [%s] was not found", filePath))
}

func expectFileNotInstalled(filePath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("stat %s", filePath))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), fmt.Sprintf("file [%s] was found when it should not have", filePath))
}

func scp(localPath string, remotePath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "scp", localPath, "mapfs:" + remotePath)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
}

func installDpkg(dpkgPath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("sudo dpkg -i %s", dpkgPath))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
