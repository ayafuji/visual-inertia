package control

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"math"
	"os/exec"
)

const (
	AS_BIN_PATH        = "/usr/bin/osascript"
	PRINT_SCRIPT_PATH  = "as/print.scpt"
	HAND_SCRIPT_PATH   = "as/hand.scpt"
	ERASER_SCRIPT_PATH = "as/eraser.scpt"
	MERGE_SCRIPT_PATH  = "as/merge.scpt"
	AT_TIME = 0
	SX_INDEX     = 1
	SY_INDEX     = 2
	SZ_INDEX     = 3
	ERASER_INDEX       = 4
	HAND_INDEX         = 5
	MERGE_INDEX        = 6
	RATIO_INDEX        = 7
	VOLUME_INDEX           = 8
	DOWN_INDEX         = 9
	DRAWALABLE_MERGIN  = 10

	DRAWABLE_AREA_WIDTH = 576
	DRAWABLE_AREA_HEIGHT = 768
)

type PCController struct {
	disable bool
}

func NewPCController() *PCController {
	pcc := PCController{disable: false}
	return &pcc
}

func (p *PCController) ToggleDisable() {
	p.disable = !p.disable
}

func (p *PCController) Print() error {
	if p.disable {
		return nil
	}
	err := exec.Command(AS_BIN_PATH, PRINT_SCRIPT_PATH).Start()
	if err != nil {
		fmt.Printf("failed to exec print command %s\n", err)
		return err
	}
	return nil
}

func (p *PCController) Merge() error {
	if p.disable {
		return nil
	}
	err := exec.Command(AS_BIN_PATH, MERGE_SCRIPT_PATH).Start()
	if err != nil {
		fmt.Printf("failed to exec merge command %s\n", err)
		return err
	}
	return nil
}

func (p *PCController) ChangeTool(tool int) error {

	toolScriptPath := ""
	if tool == ERASER_INDEX {
		fmt.Println("change to ERASER tool")
		toolScriptPath = ERASER_SCRIPT_PATH
	} else if tool == HAND_INDEX {
		fmt.Println("change to HAND tool")
		toolScriptPath = HAND_SCRIPT_PATH
	} else {
		return fmt.Errorf("invalid script number %d", tool)
	}

	if p.disable {
		return nil
	}
	// exec eraser script
	err := exec.Command(AS_BIN_PATH, toolScriptPath).Run()
	if err != nil {
		fmt.Printf("failed to exec eraser command %s\n", err)
		return err
	}
	return nil
}

func (p *PCController) MouseDrag(nx, ny int) {
	if p.disable {
		return
	}
	x, y := robotgo.GetMousePos()
	w, h := robotgo.GetScreenSize()
	if math.Abs(float64(x-nx)) < 500 &&
		math.Abs(float64(y-ny)) < 500 &&
		nx > DRAWALABLE_MERGIN &&
		nx < w-DRAWALABLE_MERGIN &&
		ny > DRAWALABLE_MERGIN &&
		ny < h-DRAWALABLE_MERGIN {
		robotgo.DragSmooth(nx, ny, 0.5)
	} else {
		fmt.Printf("invalid request, here is not drawalable %d, %d\n", nx, ny)
	}
}

func (p *PCController) MouseMove(nx, ny int) {
	if p.disable {
		return
	}
	w, h := robotgo.GetScreenSize()
	if nx > DRAWALABLE_MERGIN &&
		nx < w-DRAWALABLE_MERGIN &&
		ny > DRAWALABLE_MERGIN &&
		ny < h-DRAWALABLE_MERGIN {
		robotgo.Move(nx, ny)
		//robotgo.MoveSmooth(nx, ny, 0.5)
	}
}
