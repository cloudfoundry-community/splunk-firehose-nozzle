package cache_test

import (
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	var (
		boltdbPath         = "/tmp/boltdb"
		ignoreMissingApps  = true
		appCacheTTL        = 2 * time.Second
		missingAppCacheTTL = 2 * time.Second
		n                  = 10

		nilApp *App = nil

		config = &BoltdbConfig{
			Path:               boltdbPath,
			IgnoreMissingApps:  ignoreMissingApps,
			AppCacheTTL:        appCacheTTL,
			MissingAppCacheTTL: missingAppCacheTTL,
			Logger:             lager.NewLogger("test"),
		}

		client *testing.AppClientMock = nil
		cache  *Boltdb                = nil
		gerr   error                  = nil
	)

	BeforeEach(func() {
		os.Remove(boltdbPath)
		client = testing.NewAppClientMock(n)
		cache, gerr = NewBoltdb(client, config)
		Ω(gerr).ShouldNot(HaveOccurred())

		gerr = cache.Open()
		Ω(gerr).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		gerr = cache.Close()
		Ω(gerr).ShouldNot(HaveOccurred())

		time.Sleep(1 * time.Second)
		gerr = os.Remove(boltdbPath)
		Ω(gerr).ShouldNot(HaveOccurred())
	})

	Context("Get app good case", func() {
		It("Have 10 apps", func() {
			apps, err := cache.GetAllApps()
			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n))
		})

		It("Expect app", func() {
			guid := "cf_app_id_0"
			app, err := cache.GetApp(guid)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(app).NotTo(Equal(nil))
			Expect(app.Guid).To(Equal(guid))
		})
	})

	Context("Get app bad case", func() {
		It("Expect no app", func() {
			guid := fmt.Sprintf("cf_app_id_not_exists_%d", time.Now().UnixNano())
			app, err := cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(app).To(Equal(nilApp))

			// We ignore missing apps, so for the second time query, we already
			// recorded the missing app, so nil, err is expected to return
			app, err = cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(err).To(Equal(MissingAndIgnoredErr))
			Expect(app).To(Equal(nilApp))

			time.Sleep(missingAppCacheTTL + 2)

			// We ignore missing apps, so for the 3rd time query after sleep,
			// the missing app cache will be cleaned up, so a not found error
			// will be returned instead of MissingAndIgnoredErr
			app, err = cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(err).NotTo(Equal(MissingAndIgnoredErr))
			Expect(app).To(Equal(nilApp))
		})
	})

	Context("Cache invalidation", func() {
		It("Expect new app", func() {
			id := fmt.Sprintf("id_%d", time.Now().UnixNano())
			client.CreateApp(id, id, id)

			// Sleep for AppCacheTTL interval to make sure the cache
			// invalidation happens
			time.Sleep(appCacheTTL + 1)

			app, err := cache.GetApp(id)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(app).NotTo(Equal(nilApp))
			Expect(app.Guid).To(Equal(id))

			apps, err := cache.GetAllApps()
			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n + 1))
		})
	})

	Context("NewBoltdb error", func() {
		It("Expect error", func() {
			dup := *config
			dup.Path = fmt.Sprintf("/not-exists-%d/boltdb", time.Now().UnixNano())
			bcache, err := NewBoltdb(client, &dup)
			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("Load from existing boltdb", func() {
		It("Expect 10 apps from existing boltdb", func() {
			dup := *config
			dup.Path = fmt.Sprintf("/tmp/%d", time.Now().UnixNano())
			bcache, err := NewBoltdb(client, &dup)
			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).ShouldNot(HaveOccurred())

			defer os.Remove(dup.Path)
			time.Sleep(time.Second)
			bcache.Close()

			// Load from existing db
			bcache, err = NewBoltdb(client, &dup)
			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).ShouldNot(HaveOccurred())

			apps, err := bcache.GetAllApps()
			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n))
		})
	})
})
