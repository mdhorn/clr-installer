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
	safeButton        *gtk.RadioButton
	destructiveButton *gtk.RadioButton
	errorMessage      *gtk.Label
	rescanButton      *gtk.RadioButton
	gpartedButton     *gtk.RadioButton
	safeStore         *gtk.ListStore
	destructiveStore  *gtk.ListStore
	safeCombo         *gtk.ComboBox
	destructiveCombo  *gtk.ComboBox
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
		disk.safeCombo.SetSensitive(disk.safeButton.GetActive())
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

	log.Debug("Before making ComboBox")
	disk.safeCombo, err = gtk.ComboBoxNew()
	if err != nil {
		log.Warning("Failed to make disk.safeCombo")
		return nil, err
	}

	safeBox.PackStart(disk.safeCombo, true, true, 0)

	log.Debug("Before safeBox ShowAll")
	safeBox.ShowAll()
	disk.scrollBox.Add(safeBox)

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
		disk.destructiveCombo.SetSensitive(disk.destructiveButton.GetActive())
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

	log.Debug("Before making ComboBox")
	disk.destructiveCombo, err = gtk.ComboBoxNew()
	if err != nil {
		log.Warning("Failed to make disk.destructiveCombo")
		return nil, err
	}

	destructiveBox.PackStart(disk.destructiveCombo, true, true, 0)

	destructiveBox.ShowAll()
	disk.scrollBox.Add(destructiveBox)

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

	// Build the Rescan Section
	disk.rescanButton, err = gtk.RadioButtonNewFromWidget(disk.destructiveButton)
	if err != nil {
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
	disk.gpartedButton, err = gtk.RadioButtonNewFromWidget(disk.rescanButton)
	if err != nil {
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
	safeStore, err := gtk.ListStoreNew(glib.TYPE_STRING)
	if err != nil {
		log.Warning("ListStoreNew failed")
		return err
	}
	destructiveStore, err := gtk.ListStoreNew(glib.TYPE_STRING)
	if err != nil {
		log.Warning("ListStoreNew failed")
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
				err := safeStore.SetValue(safeStore.Append(), 0, target.Name)
				if err != nil {
					log.Warning("SetValue safeStore")
					return err
				}
				safeFound = true
				break
			}
		}
		if !found {
			log.Debug("Adding destructive install target %s", device.Name)
			err := destructiveStore.SetValue(destructiveStore.Append(), 0, device.Name)
			if err != nil {
				log.Warning("SetValue destructiveStore")
				return err
			}
			destructiveFound = true
		}
	}

	disk.safeCombo.SetModel(safeStore)
	cellRenderer, _ := gtk.CellRendererTextNew()
	disk.safeCombo.PackStart(cellRenderer, true)
	disk.safeCombo.AddAttribute(cellRenderer, "text", 0)
	if safeFound {
		disk.safeCombo.SetActive(0)
	}

	disk.destructiveCombo.SetModel(destructiveStore)
	cellRenderer2, _ := gtk.CellRendererTextNew()
	disk.destructiveCombo.PackStart(cellRenderer2, true)
	disk.destructiveCombo.AddAttribute(cellRenderer2, "text", 0)
	if destructiveFound {
		disk.destructiveCombo.SetActive(0)
	}

	if _, err := disk.box.Connect("show", func() {
		log.Debug("We triggered a visibility-notify-event")
		/*
			if err := disk.populateComboBoxes(); err != nil {
				log.Warning("Problem building Button Section for disk selection")
			}
		*/
	}); err != nil {
		return nil
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

	disk.safeCombo.SetSensitive(false)
	disk.destructiveCombo.SetSensitive(false)

	if err := disk.populateComboBoxes(); err != nil {
		log.Warning("Problem populating possible disk selections")
	}

	// Choose the most appropriate button
	if len(disk.installTargets) > 0 {
		disk.safeButton.SetActive(true)
		disk.safeCombo.SetSensitive(true)
	} else if len(disk.devs) > 0 {
		disk.destructiveButton.SetActive(true)
		disk.destructiveCombo.SetSensitive(true)
	} else {
		disk.rescanButton.SetActive(true)
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
