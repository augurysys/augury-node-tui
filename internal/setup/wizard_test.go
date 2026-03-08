package setup

import (
	"strings"
	"testing"
)

func TestWizard_InitStartsAtRootStep(t *testing.T) {
	w := NewWizard(false)
	if w.currentStep != 0 {
		t.Error("Wizard should start at step 0")
	}
}

func TestWizard_NextStepMsgAdvances(t *testing.T) {
	w := NewWizard(false)
	w.currentStep = 0
	model, _ := w.Update(NextStepMsg{})
	w = model.(*WizardModel)
	if w.currentStep != 1 {
		t.Errorf("NextStepMsg should advance step; got %d, want 1", w.currentStep)
	}
}

func TestWizard_ViewShowsProgressIndicator(t *testing.T) {
	w := NewWizard(false)
	w.currentStep = 2
	view := w.View()
	if !strings.Contains(view, "Step") || !strings.Contains(view, "/6") {
		t.Error("View should show step progress indicator")
	}
}

func TestWizard_LaunchMainTUIExits(t *testing.T) {
	w := NewWizard(false)
	w.currentStep = 5
	model, cmd := w.Update(LaunchMainTUIMsg{})
	w = model.(*WizardModel)
	if w.launchMain != true {
		t.Error("LaunchMainTUIMsg should set launchMain flag")
	}
	if cmd == nil {
		t.Error("Should return quit command")
	}
}
