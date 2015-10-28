package deploy_test

import (
	"acceptance-tests/helpers"

	capi "github.com/hashicorp/consul/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Scaling up Instances", func() {
	var (
		consulManifest  *helpers.Manifest
		consulServerIPs []string
		runner          *helpers.AgentRunner
	)

	BeforeEach(func() {
		consulManifest = new(helpers.Manifest)
		consulServerIPs = []string{}

		By("deploying 3 nodes")
		bosh.GenerateAndSetDeploymentManifest(
			consulManifest,
			consulManifestGeneration,
			directorUUIDStub,
			helpers.InstanceCount3NodesStubPath,
			helpers.PersistentDiskStubPath,
			config.IAASSettingsConsulStubPath,
			helpers.PropertyOverridesStubPath,
			consulNameOverrideStub,
		)

		Expect(bosh.Command("-n", "deploy")).To(gexec.Exit(0))
		Expect(len(consulManifest.Properties.Consul.Agent.Servers.Lans)).To(Equal(3))

		for _, elem := range consulManifest.Properties.Consul.Agent.Servers.Lans {
			consulServerIPs = append(consulServerIPs, elem)
		}

		runner = helpers.NewAgentRunner(consulServerIPs, config.BindAddress)
		runner.Start()
	})

	AfterEach(func() {
		// By("delete deployment")
		// runner.Stop()
		// bosh.Command("-n", "delete", "deployment", consulDeployment)
	})

	Describe("scaling from 3 nodes to 1", func() {
		FIt("succesfully scales from multiple consul nodes to one consul node without saving data", func() {

			By("scaling down to 1 node")
			bosh.GenerateAndSetDeploymentManifest(
				consulManifest,
				consulManifestGeneration,
				directorUUIDStub,
				helpers.InstanceCount1NodeStubPath,
				helpers.PersistentDiskStubPath,
				config.IAASSettingsConsulStubPath,
				helpers.PropertyOverridesStubPath,
				consulNameOverrideStub,
			)

			By("deploying")
			Expect(bosh.Command("-n", "deploy")).To(gexec.Exit(0))
			Expect(len(consulManifest.Properties.Consul.Agent.Servers.Lans)).To(Equal(1))

			By("setting and reading a value from consul")
			consatsClient := runner.NewClient()

			consatsKey := "consats-key"
			consatsValue := []byte("consats-value")

			keyValueClient := consatsClient.KV()

			pair := &capi.KVPair{Key: consatsKey, Value: consatsValue}

			Eventually(func() error {
				_, err := keyValueClient.Put(pair, nil)
				return err
			}, "5m", "10s").ShouldNot(HaveOccurred())

			resultPair, _, err := keyValueClient.Get(consatsKey, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(resultPair.Value).To(Equal(consatsValue))
			//Expect(resultPair).NotTo(BeNil())
			//Expect(resultPair.Value).To(Equal(consatsValue))

			/*
				By("scaling down to 1 node")
				bosh.GenerateAndSetDeploymentManifest(
					consulManifest,
					consulManifestGeneration,
					directorUUIDStub,
					helpers.InstanceCount1NodeStubPath,
					helpers.PersistentDiskStubPath,
					config.IAASSettingsConsulStubPath,
					helpers.PropertyOverridesStubPath,
					consulNameOverrideStub,
				)

				Expect(bosh.Command("-n", "deploy")).To(gexec.Exit(0))
				Expect(len(consulManifest.Properties.Consul.Agent.Servers.Lans)).To(Equal(1))

				By("setting a value")

				consatsKey := "consats-key"
				consatsValue := []byte("consats-value")
				pair := &capi.KVPair{Key: consatsKey, Value: consatsValue}

				consatsClient := runner.NewClient()
				keyValueClient := consatsClient.KV()
				_, err := keyValueClient.Put(pair, nil)
				Expect(err).ToNot(HaveOccurred())

				resultPair, _, err := keyValueClient.Get(consatsKey, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(resultPair).NotTo(BeNil())
				Expect(resultPair.Value).To(Equal(consatsValue))
			*/
		})
	})
})
