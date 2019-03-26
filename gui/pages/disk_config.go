// Copyright © 2019 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package pages

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/clearlinux/clr-installer/log"
	"github.com/clearlinux/clr-installer/model"
	"github.com/clearlinux/clr-installer/storage"
)

// DiskConfig is a simple page to help with DiskConfig settings
type DiskConfig struct {
	devs              []*storage.BlockDevice
	installTargets    []storage.InstallTarget
	activeDisk        *storage.BlockDevice
	controller        Controller
	model             *model.SystemInstall
	box               *gtk.Box
	scroll            *gtk.ScrolledWindow
	scrollBox         *gtk.Box
	mediaGrid         *gtk.Grid
	safeButton        *gtk.RadioButton
	destructiveButton *gtk.RadioButton
	chooserCombo      *gtk.ComboBox
	errorMessage      *gtk.Label
	rescanButton      *gtk.Button
	gpartedButton     *gtk.Button
}

func newListStoreMedia() (*gtk.ListStore, error) {
	store, err := gtk.ListStoreNew(glib.TYPE_OBJECT, glib.TYPE_STRING, glib.TYPE_STRING)
	return store, err
}

// addListStoreMediaRow adds new row to the ListStore widget for the given media
func addListStoreMediaRow(store *gtk.ListStore, installMedia storage.InstallTarget) error {

	// Create icon image
	mediaType := "drive-harddisk-system"
	if installMedia.Removable {
		mediaType = "media-removable"
	}
	mediaType = mediaType + "-symbolic"
	image, err := gtk.ImageNewFromIconName(mediaType, gtk.ICON_SIZE_DIALOG)
	if err != nil {
		log.Warning("gtk.ImageNewFromIconName failed for icon %q", mediaType)
		return err
	}

	iter := store.Append()

	err = store.SetValue(iter, 0, image.GetPixbuf())
	if err != nil {
		log.Warning("SetValue store failed for icon %q", mediaType)
		return err
	}

	// Name string
	nameString := installMedia.Friendly

	err = store.SetValue(iter, 1, nameString)
	if err != nil {
		log.Warning("SetValue store failed for name string: %q", nameString)
		return err
	}

	// Size string
	sizeString := "[Partial]"
	if installMedia.WholeDisk {
		sizeString = "[Full Disk]"
	}

	err = store.SetValue(iter, 2, sizeString)
	if err != nil {
		log.Warning("SetValue store failed for size string: %q", sizeString)
		return err
	}

	return nil
}

// NewDiskConfigPage returns a new DiskConfigPage
func NewDiskConfigPage(controller Controller, model *model.SystemInstall) (Page, error) {
	disk := &DiskConfig{
		controller: controller,
		model:      model,
	}
	var err error

	disk.box, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}
	disk.box.SetBorderWidth(8)

	// Build storage for scrollBox
	disk.scroll, err = gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	disk.box.PackStart(disk.scroll, true, true, 0)
	disk.scroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)

	// Build scrollBox
	disk.scrollBox, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 20)
	if err != nil {
		return nil, err
	}

	disk.scroll.Add(disk.scrollBox)

	// Media Grid
	disk.mediaGrid, err = gtk.GridNew()
	if err != nil {
		return nil, err
	}

	// Build the Safe Install Section
	disk.safeButton, err = gtk.RadioButtonNewFromWidget(nil)
	if err != nil {
		return nil, err
	}

	safeBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	safeBox.PackStart(disk.safeButton, false, false, 10)
	disk.safeButton.Connect("toggled", func() {
		// Enable/Disable the Combo Choose Box based on the radio button
		//disk.safeCombo.SetSensitive(disk.safeButton.GetActive())
	})

	safeHortzBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	safeBox.PackStart(safeHortzBox, true, true, 0)
	text := fmt.Sprintf("<big>Safe Installation</big>\n")
	text = text + "Install on available media without the loss of data"
	safeLabel, err := gtk.LabelNew(text)
	if err != nil {
		return nil, err
	}
	safeLabel.SetXAlign(0.0)
	safeLabel.SetHAlign(gtk.ALIGN_START)
	safeLabel.SetUseMarkup(true)
	safeHortzBox.PackStart(safeLabel, false, false, 0)

	log.Debug("Before safeBox ShowAll")
	safeBox.ShowAll()
	disk.mediaGrid.Attach(safeBox, 0, 0, 1, 1)

	// Build the Destructive Install Section
	log.Debug("Before disk.destructiveButton")
	disk.destructiveButton, err = gtk.RadioButtonNewFromWidget(disk.safeButton)
	if err != nil {
		return nil, err
	}

	destructiveBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	destructiveBox.PackStart(disk.destructiveButton, false, false, 10)
	disk.destructiveButton.Connect("toggled", func() {
		// Enable/Disable the Combo Choose Box based on the radio button
		//disk.destructiveCombo.SetSensitive(disk.destructiveButton.GetActive())
	})

	destructiveHortzBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	destructiveBox.PackStart(destructiveHortzBox, true, true, 0)
	text = fmt.Sprintf("<big><b><span foreground=\"red\">Destructive Installation</span></b></big>\n")
	text = text + "Install on available media wiping all data from media!!"
	destructiveLabel, err := gtk.LabelNew(text)
	if err != nil {
		return nil, err
	}
	destructiveLabel.SetXAlign(0.0)
	destructiveLabel.SetHAlign(gtk.ALIGN_START)
	destructiveLabel.SetUseMarkup(true)
	destructiveHortzBox.PackStart(destructiveLabel, false, false, 0)

	destructiveBox.ShowAll()
	disk.mediaGrid.Attach(destructiveBox, 0, 1, 1, 1)

	log.Debug("Before making ComboBox")
	disk.chooserCombo, err = gtk.ComboBoxNew()
	if err != nil {
		log.Warning("Failed to make disk.chooserCombo")
		return nil, err
	}

	// Add the renderers to the ComboBox
	mediaRenderer, _ := gtk.CellRendererPixbufNew()
	disk.chooserCombo.PackStart(mediaRenderer, true)
	disk.chooserCombo.AddAttribute(mediaRenderer, "pixbuf", 0)

	nameRenderer, _ := gtk.CellRendererTextNew()
	disk.chooserCombo.PackStart(nameRenderer, true)
	disk.chooserCombo.AddAttribute(nameRenderer, "text", 1)

	sizeRenderer, _ := gtk.CellRendererTextNew()
	disk.chooserCombo.PackStart(sizeRenderer, true)
	disk.chooserCombo.AddAttribute(sizeRenderer, "text", 2)

	disk.mediaGrid.Attach(disk.chooserCombo, 1, 0, 1, 2)

	disk.mediaGrid.SetRowSpacing(10)
	disk.mediaGrid.SetColumnSpacing(10)
	disk.mediaGrid.SetColumnHomogeneous(true)
	disk.scrollBox.Add(disk.mediaGrid)

	separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	separator.ShowAll()
	disk.scrollBox.Add(separator)

	// Error Message Label
	disk.errorMessage, err = gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	disk.errorMessage.SetXAlign(0.0)
	disk.errorMessage.SetHAlign(gtk.ALIGN_START)
	disk.errorMessage.SetUseMarkup(true)
	disk.scrollBox.Add(disk.errorMessage)

	// Build the Rescan Button
	if disk.rescanButton, err = createNavButton("RESCAN"); err != nil {
		return nil, err
	}
	if _, err = disk.rescanButton.Connect("clicked", func() {
		log.Debug("rescan")
		_ = disk.scanMediaDevices()
		if err := disk.populateComboBoxes(); err != nil {
			log.Warning("Problem populating possible disk selections")
		}
	}); err != nil {
		return nil, err
	}

	rescanBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	rescanBox.PackStart(disk.rescanButton, false, false, 10)

	rescanHortzBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	rescanBox.PackStart(rescanHortzBox, true, true, 0)
	text = fmt.Sprintf("<big>Rescan Media</big>\n")
	text = text + "Rescan for changes to hot swappable media."
	rescanLabel, err := gtk.LabelNew(text)
	if err != nil {
		return nil, err
	}
	rescanLabel.SetXAlign(0.0)
	rescanLabel.SetHAlign(gtk.ALIGN_START)
	rescanLabel.SetUseMarkup(true)
	rescanHortzBox.PackStart(rescanLabel, false, false, 0)

	rescanBox.ShowAll()
	disk.scrollBox.Add(rescanBox)

	// Build the Gparted Section
	if disk.gpartedButton, err = createNavButton("GPARTED"); err != nil {
		return nil, err
	}
	if _, err = disk.gpartedButton.Connect("clicked", func() {
		log.Debug("launching gparted")
		//TODO: launch external gui app
		// Can we get this program to 'pause' until gparted exists?
	}); err != nil {
		return nil, err
	}

	gpartedBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	gpartedBox.PackStart(disk.gpartedButton, false, false, 10)

	gpartedHortzBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	gpartedBox.PackStart(gpartedHortzBox, true, true, 0)
	text = fmt.Sprintf("<big>Manual Update Disk</big>\n")
	text = text + "Modify disk to be usable for installation.\n"
	text = text + "We need at least <b>20GB</b> of space.\n"
	gpartedLabel, err := gtk.LabelNew(text)
	if err != nil {
		return nil, err
	}
	gpartedLabel.SetXAlign(0.0)
	gpartedLabel.SetHAlign(gtk.ALIGN_START)
	gpartedLabel.SetUseMarkup(true)
	gpartedHortzBox.PackStart(gpartedLabel, false, false, 0)

	gpartedBox.ShowAll()
	disk.scrollBox.Add(gpartedBox)

	disk.box.ShowAll()

	_ = disk.scanMediaDevices()

	return disk, nil
}

// This is time intensive, mitigate calls
func (disk *DiskConfig) scanMediaDevices() error {
	var err error

	disk.devs, err = storage.RescanBlockDevices(nil)
	if err != nil {
		return err
	}

	disk.installTargets = storage.FindInstallTargets(storage.MinimumInstallSize, disk.devs)

	return nil
}

// populateComboBoxes populates the scrollBox with usable widget things
func (disk *DiskConfig) populateComboBoxes() error {
	isSafe := true

	safeStore, err := newListStoreMedia()
	if err != nil {
		log.Warning("ListStoreNew safeStore failed")
		return err
	}
	destructiveStore, err := newListStoreMedia()
	if err != nil {
		log.Warning("ListStoreNew destructiveStore failed")
		return err
	}

	safeFound := false
	destructiveFound := false

	if len(disk.devs) < 1 {
		warning := "No media found for installation"
		log.Warning(warning)
		warning = fmt.Sprintf("<big><b><span foreground=\"red\">Warning: %s</span></b></big>", warning)
		disk.errorMessage.SetMarkup(warning)
		return nil
	}

	for _, device := range disk.devs {
		found := false
		for _, target := range disk.installTargets {
			if device.Name == target.Name {
				found = true
				log.Debug("Adding safe install target %s", target.Name)
				err := addListStoreMediaRow(safeStore, target)
				if err != nil {
					log.Warning("SetValue safeStore")
					return err
				}
				safeFound = true
				break
			}
		}
		if !found {
			target := storage.InstallTarget{Name: device.Name, Friendly: device.Model,
				WholeDisk: true, Removable: device.RemovableDevice}
			log.Debug("Adding destructive install target %s", target.Name)
			err := addListStoreMediaRow(destructiveStore, target)
			if err != nil {
				log.Warning("SetValue destructiveStore")
				return err
			}
			destructiveFound = true
		}
	}

	if isSafe {
		disk.chooserCombo.SetModel(safeStore)
		if safeFound {
			disk.chooserCombo.SetActive(0)
		}
	} else {
		disk.chooserCombo.SetModel(destructiveStore)
		if destructiveFound {
			disk.chooserCombo.SetActive(0)
		}
	}

	return nil
}

// Set the right disk for installation
func (disk *DiskConfig) onRowActivated(box *gtk.ListBox, row *gtk.ListBoxRow) {
	if row == nil {
		disk.activeDisk = nil
		disk.controller.SetButtonState(ButtonConfirm, false)
		return
	}
	disk.controller.SetButtonState(ButtonConfirm, true)
	idx := row.GetIndex()
	log.Debug("We just selected row %d", idx)
	//disk.activeDisk = disk.devs[idx]
}

// IsRequired will return true as we always need a DiskConfig
func (disk *DiskConfig) IsRequired() bool {
	return true
}

// IsDone checks if all the steps are completed
func (disk *DiskConfig) IsDone() bool {
	return disk.model.TargetMedias != nil
}

// GetID returns the ID for this page
func (disk *DiskConfig) GetID() int {
	return PageIDDiskConfig
}

// GetIcon returns the icon for this page
func (disk *DiskConfig) GetIcon() string {
	return "drive-harddisk-system"
}

// GetRootWidget returns the root embeddable widget for this page
func (disk *DiskConfig) GetRootWidget() gtk.IWidget {
	return disk.box
}

// GetSummary will return the summary for this page
func (disk *DiskConfig) GetSummary() string {
	return "Configure Media"
}

// GetTitle will return the title for this page
func (disk *DiskConfig) GetTitle() string {
	return disk.GetSummary() + " - WARNING: SUPER EXPERIMENTAL"
}

// StoreChanges will store this pages changes into the model
func (disk *DiskConfig) StoreChanges() {
	// Give the active disk to the model
	disk.model.AddTargetMedia(disk.activeDisk)
}

// ResetChanges will reset this page to match the model
func (disk *DiskConfig) ResetChanges() {
	disk.activeDisk = nil
	disk.controller.SetButtonState(ButtonConfirm, true)

	disk.chooserCombo.SetSensitive(false)

	if err := disk.populateComboBoxes(); err != nil {
		log.Warning("Problem populating possible disk selections")
	}

	// Choose the most appropriate button
	if len(disk.installTargets) > 0 {
		disk.safeButton.SetActive(true)
		disk.chooserCombo.SetSensitive(true)
	} else if len(disk.devs) > 0 {
		disk.destructiveButton.SetActive(true)
		disk.chooserCombo.SetSensitive(true)
	} else {
		//disk.rescanButton.SetActive(true)
		//TODO: Make this button have focus/default
		log.Debug("Need to make the rescan button default")
	}

	// TODO: Match list to target medias. But we have an ugly
	// list of root target medias and you can only select one
	// right now as our manual partitioning is missing.
	if disk.model.TargetMedias == nil {
		return
	}
}

// GetConfiguredValue returns our current config
func (disk *DiskConfig) GetConfiguredValue() string {
	if disk.model.TargetMedias == nil {
		return "No usable media found"
	}
	return fmt.Sprintf("WARNING: Wiping %s", disk.model.TargetMedias[0].GetDeviceFile())
}

// createNavButton creates specialised navigation button
func createNavButton(label string) (*gtk.Button, error) {
	var st *gtk.StyleContext
	button, err := gtk.ButtonNewWithLabel(label)
	if err != nil {
		return nil, err
	}

	st, err = button.GetStyleContext()
	if err != nil {
		return nil, err
	}
	st.AddClass("nav-button")
	return button, nil
}
