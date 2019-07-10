package bosh_release_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)

var _ = Describe("BoshReleaseTest", func() {
	BeforeEach(func() {
		deploy()
	})

	It("should have the mapfs binaries", func() {
		expectFileInstalled("/var/vcap/packages/mapfs/bin/mapfs")
		expectDpkgInstalled(" fuse ")
		expectFileInstalled("/etc/fuse.conf")
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
})

func expectDpkgNotInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "dpkg -l | grep '%s'", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), fmt.Sprintf("package [%s] was found when it should not have.", dpkgName))
}

func expectDpkgInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "dpkg -l | grep '%s'", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), fmt.Sprintf("package [%s] was not found when it should have.", dpkgName))
}

func expectFileInstalled(filePath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "stat %s", filePath))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), fmt.Sprintf("file [%s] was not found", filePath))
}

func expectFileNotInstalled(filePath string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "stat %s", filePath))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), fmt.Sprintf("file [%s] was found when it should not have", filePath))
}
