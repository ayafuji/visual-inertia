package control

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"os/exec"
)

const (
	HAND_TOOL = 0
	ERASER_TOOL = 1
	AS_BIN_PATH = "/usr/bin/osascript"
	PRINT_SCRIPT_PATH = "as/print.scpt"
	HAND_SCRIPT_PATH = "as/hand.scpt"
	ERASER_SCRIPT_PATH = "as/eraser.scpt"
	MERGE_SCRIPT_PATH = "as/merge.scpt"
	ERASER_INDEX = 3
	HAND_INDEX = 4
	MERGE_INDEX = 5
	PRINT_INDEX = 6
)

type PCController struct {

}

func NewPCController() *PCController {
	pcc := PCController{}
	return &pcc
}

func (p *PCController) Print() error {
	err := exec.Command(AS_BIN_PATH, PRINT_SCRIPT_PATH).Start()
	if err != nil {
		fmt.Printf("failed to exec print command %s\n", err)
		return err
	}
	return nil
}

func (p *PCController) Merge() error {
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
	// exec eraser script
	err := exec.Command(AS_BIN_PATH, toolScriptPath).Run()
	if err != nil {
		fmt.Printf("failed to exec eraser command %s\n", err)
		return err
	}
	return nil
}

func (p *PCController) MouseDrag(nx, ny int) {
	robotgo.DragSmooth(nx, ny,  0.5)
}

func (p *PCController) MouseMove(nx, ny int) {
	robotgo.MoveSmooth(nx, ny,  0.5)
}


