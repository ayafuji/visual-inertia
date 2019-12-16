package network

import "github.com/hypebeast/go-osc/osc"

func SendOSCFloat(client *osc.Client, value float32, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}

func SendOSCInt(client *osc.Client, value int32, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}

func SendOSCString(client *osc.Client, value string, path string) error {
	message := osc.NewMessage(path)
	message.Append(value)
	if err := client.Send(message); err != nil {
		return err
	}
	return nil
}
