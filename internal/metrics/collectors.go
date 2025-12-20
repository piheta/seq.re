package metrics

import (
	"github.com/piheta/seq.re/internal/features/img"
	"github.com/piheta/seq.re/internal/features/link"
	"github.com/piheta/seq.re/internal/features/paste"
	"github.com/piheta/seq.re/internal/features/secret"
	"github.com/prometheus/client_golang/prometheus"
)

type LinkCollector struct {
	linkRepo         *link.LinkRepo
	encryptedLinks   *prometheus.Desc
	unencryptedLinks *prometheus.Desc
}

func NewLinkCollector(linkRepo *link.LinkRepo) *LinkCollector {
	return &LinkCollector{
		linkRepo: linkRepo,
		encryptedLinks: prometheus.NewDesc(
			"seqre_links_encrypted_total",
			"Total number of encrypted links in the database",
			nil,
			nil,
		),
		unencryptedLinks: prometheus.NewDesc(
			"seqre_links_unencrypted_total",
			"Total number of unencrypted links in the database",
			nil,
			nil,
		),
	}
}

func (c *LinkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.encryptedLinks
	ch <- c.unencryptedLinks
}

func (c *LinkCollector) Collect(ch chan<- prometheus.Metric) {
	encrypted, unencrypted, err := c.linkRepo.CountLinks()
	if err != nil {
		// If there's an error, report 0 for both metrics
		ch <- prometheus.MustNewConstMetric(c.encryptedLinks, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.unencryptedLinks, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.encryptedLinks, prometheus.GaugeValue, float64(encrypted))
	ch <- prometheus.MustNewConstMetric(c.unencryptedLinks, prometheus.GaugeValue, float64(unencrypted))
}

type ImageCollector struct {
	imageRepo         *img.ImageRepo
	encryptedImages   *prometheus.Desc
	unencryptedImages *prometheus.Desc
}

func NewImageCollector(imageRepo *img.ImageRepo) *ImageCollector {
	return &ImageCollector{
		imageRepo: imageRepo,
		encryptedImages: prometheus.NewDesc(
			"seqre_images_encrypted_total",
			"Total number of encrypted images in the database",
			nil,
			nil,
		),
		unencryptedImages: prometheus.NewDesc(
			"seqre_images_unencrypted_total",
			"Total number of unencrypted images in the database",
			nil,
			nil,
		),
	}
}

func (c *ImageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.encryptedImages
	ch <- c.unencryptedImages
}

func (c *ImageCollector) Collect(ch chan<- prometheus.Metric) {
	encrypted, unencrypted, err := c.imageRepo.CountImages()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.encryptedImages, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.unencryptedImages, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.encryptedImages, prometheus.GaugeValue, float64(encrypted))
	ch <- prometheus.MustNewConstMetric(c.unencryptedImages, prometheus.GaugeValue, float64(unencrypted))
}

type PasteCollector struct {
	pasteRepo         *paste.PasteRepo
	encryptedPastes   *prometheus.Desc
	unencryptedPastes *prometheus.Desc
}

func NewPasteCollector(pasteRepo *paste.PasteRepo) *PasteCollector {
	return &PasteCollector{
		pasteRepo: pasteRepo,
		encryptedPastes: prometheus.NewDesc(
			"seqre_pastes_encrypted_total",
			"Total number of encrypted pastes in the database",
			nil,
			nil,
		),
		unencryptedPastes: prometheus.NewDesc(
			"seqre_pastes_unencrypted_total",
			"Total number of unencrypted pastes in the database",
			nil,
			nil,
		),
	}
}

func (c *PasteCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.encryptedPastes
	ch <- c.unencryptedPastes
}

func (c *PasteCollector) Collect(ch chan<- prometheus.Metric) {
	encrypted, unencrypted, err := c.pasteRepo.CountPastes()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.encryptedPastes, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.unencryptedPastes, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.encryptedPastes, prometheus.GaugeValue, float64(encrypted))
	ch <- prometheus.MustNewConstMetric(c.unencryptedPastes, prometheus.GaugeValue, float64(unencrypted))
}

type SecretCollector struct {
	secretRepo   *secret.SecretRepo
	totalSecrets *prometheus.Desc
}

func NewSecretCollector(secretRepo *secret.SecretRepo) *SecretCollector {
	return &SecretCollector{
		secretRepo: secretRepo,
		totalSecrets: prometheus.NewDesc(
			"seqre_secrets_total",
			"Total number of secrets in the database (all secrets are encrypted)",
			nil,
			nil,
		),
	}
}

func (c *SecretCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalSecrets
}

func (c *SecretCollector) Collect(ch chan<- prometheus.Metric) {
	total, err := c.secretRepo.CountSecrets()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.totalSecrets, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.totalSecrets, prometheus.GaugeValue, float64(total))
}
