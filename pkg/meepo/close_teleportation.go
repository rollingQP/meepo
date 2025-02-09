package meepo

import (
	"fmt"

	"github.com/PeerXu/meepo/pkg/transport"
	"github.com/sirupsen/logrus"
)

type CloseTeleportationRequest struct {
	*Message

	Name string `json:"name"`
}

type CloseTeleportationResponse struct {
	*Message
}

func (mp *Meepo) CloseTeleportation(name string) error {
	var err error

	logger := mp.getLogger().WithFields(logrus.Fields{
		"#method": "CloseTeleportation",
		"name":    name,
	})

	tp, err := mp.GetTeleportation(name, WithSourceFirst())
	if err != nil {
		logger.WithError(err).Errorf("failed to get teleportation")
		return err
	}

	req := &CloseTeleportationRequest{
		Message: mp.createRequest("closeTeleportation"),
		Name:    name,
	}

	out, err := mp.doRequest(tp.Transport().PeerID(), req)
	if err != nil {
		logger.WithError(err).Errorf("failed to do request")
		return err
	}
	res := out.(*CloseTeleportationResponse)
	if res.Error != "" {
		err = fmt.Errorf(res.Error)
		logger.WithError(err).Errorf("failed to close teleportation by peer")
		return err
	}

	go func() {
		if err = tp.Close(); err != nil {
			logger.WithError(err).Errorf("failed to close teleportation")
			return
		}
		logger.Infof("teleportation closed")
	}()

	logger.Debugf("teleportation closing")

	return nil
}

func (mp *Meepo) onCloseTeleportation(dc transport.DataChannel, in interface{}) {
	var err error

	req := in.(*CloseTeleportationRequest)

	logger := mp.getLogger().WithFields(
		logrus.Fields{
			"#method": "onCloseTeleportation",
			"name":    req.Name,
		})

	ts, err := mp.GetTeleportation(req.Name, WithSinkFirst())
	if err != nil {
		logger.WithError(err).Errorf("failed to get teleportation")
		mp.sendMessage(dc, mp.invertMessageWithError(req, err))
		return
	}

	go func() {
		if err = ts.Close(); err != nil {
			logger.WithError(err).Errorf("failed to close teleportation")
			return
		}
		logger.Infof("teleportation closed")
	}()

	res := CloseTeleportationResponse{
		Message: mp.invertMessage(req),
	}
	mp.sendMessage(dc, &res)

	logger.Debugf("teleportation closing")
}

func init() {
	registerDecodeMessageHelper(MESSAGE_TYPE_REQUEST, "closeTeleportation", func() interface{} { return &CloseTeleportationRequest{} })
	registerDecodeMessageHelper(MESSAGE_TYPE_RESPONSE, "closeTeleportation", func() interface{} { return &CloseTeleportationResponse{} })
}
