package conformancetest

import (
	"context"
	"time"

	"github.com/JoeReid/jetbridge/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slices"
)

type BindingsConformanceSuite struct {
	*suite.Suite

	peerID    uuid.UUID
	Peers     repositories.Peers
	Candidate repositories.Bindings
}

func (s *BindingsConformanceSuite) SetupSuite() {
	p, err := s.Peers.JoinPeers(context.TODO())
	s.Require().NoError(err)

	s.peerID = p.ID
}

func (s *BindingsConformanceSuite) SetupTest() {
	_, err := s.Peers.SendHeartbeat(context.TODO(), s.peerID)
	s.Require().NoError(err)
}

func (s *BindingsConformanceSuite) TestCreateJetstreamBinding() {
	jb, err := s.Candidate.CreateJetstreamBinding(context.TODO(), &repositories.CreateJetstreamBinding{
		LambdaARN:      "arn:aws:lambda:us-east-1:123456789012:function:my-function",
		Stream:         "my-stream",
		Subject:        "my-subject",
		MaxMessages:    0,
		MaxLatency:     0,
		DeliveryPolicy: "all",
	})
	s.Require().NoError(err)
	s.Require().NotNil(jb)

	s.Assert().NotNil(jb.ID)
	s.Assert().Equal("arn:aws:lambda:us-east-1:123456789012:function:my-function", jb.LambdaARN)
	s.Assert().Equal("my-stream", jb.Stream)
	s.Assert().Equal(jb.ID, jb.Consumer)
	s.Assert().Equal("my-subject", jb.Subject)
	s.Assert().Equal(0, jb.MaxMessages)
	s.Assert().Equal(time.Duration(0), jb.MaxLatency)
	s.Assert().Equal("all", jb.DeliveryPolicy)
	s.Assert().Equal(s.peerID, *jb.AssignedPeerID)
}

func (s *BindingsConformanceSuite) TestCreateJetstreamBinding_batched() {
	jb, err := s.Candidate.CreateJetstreamBinding(context.TODO(), &repositories.CreateJetstreamBinding{
		LambdaARN:      "arn:aws:lambda:us-east-1:123456789012:function:my-function",
		Stream:         "my-stream",
		Subject:        "my-subject",
		MaxMessages:    10,
		MaxLatency:     5 * time.Second,
		DeliveryPolicy: "all",
	})
	s.Require().NoError(err)
	s.Require().NotNil(jb)

	s.Assert().NotNil(jb.ID)
	s.Assert().Equal("arn:aws:lambda:us-east-1:123456789012:function:my-function", jb.LambdaARN)
	s.Assert().Equal("my-stream", jb.Stream)
	s.Assert().Equal(jb.ID, jb.Consumer)
	s.Assert().Equal("my-subject", jb.Subject)
	s.Assert().Equal(10, jb.MaxMessages)
	s.Assert().Equal(5*time.Second, jb.MaxLatency)
	s.Assert().Equal("all", jb.DeliveryPolicy)
	s.Assert().Equal(s.peerID, *jb.AssignedPeerID)
}

func (s *BindingsConformanceSuite) TestGetJetstreamBinding() {
	jb, err := s.Candidate.CreateJetstreamBinding(context.TODO(), &repositories.CreateJetstreamBinding{
		LambdaARN:      "arn:aws:lambda:us-east-1:123456789012:function:my-function",
		Stream:         "my-stream",
		Subject:        "my-subject",
		MaxMessages:    10,
		MaxLatency:     5 * time.Second,
		DeliveryPolicy: "all",
	})
	s.Require().NoError(err)
	s.Require().NotNil(jb)

	got, err := s.Candidate.GetJetstreamBinding(context.TODO(), jb.ID)
	s.Require().NoError(err)
	s.Require().NotNil(got)

	s.Assert().Equal(jb.ID, got.ID)
	s.Assert().Equal(jb.LambdaARN, got.LambdaARN)
	s.Assert().Equal(jb.Stream, got.Stream)
	s.Assert().Equal(jb.ID, got.Consumer)
	s.Assert().Equal(jb.Subject, got.Subject)
	s.Assert().Equal(jb.MaxMessages, got.MaxMessages)
	s.Assert().Equal(jb.MaxLatency, got.MaxLatency)
	s.Assert().Equal(jb.DeliveryPolicy, got.DeliveryPolicy)
	s.Assert().Equal(jb.AssignedPeerID, got.AssignedPeerID)
}

func (s *BindingsConformanceSuite) TestGetJetstreamBinding_notFound() {
	got, err := s.Candidate.GetJetstreamBinding(context.TODO(), uuid.New())
	s.Require().Error(err)
	s.Require().Nil(got)
}

func (s *BindingsConformanceSuite) TestListJetstreamBindings() {
	jb, err := s.Candidate.CreateJetstreamBinding(context.TODO(), &repositories.CreateJetstreamBinding{
		LambdaARN:      "arn:aws:lambda:us-east-1:123456789012:function:my-function",
		Stream:         "my-stream",
		Subject:        "my-subject",
		MaxMessages:    10,
		MaxLatency:     5 * time.Second,
		DeliveryPolicy: "all",
	})
	s.Require().NoError(err)
	s.Require().NotNil(jb)

	list, err := s.Candidate.ListJetstreamBindings(context.TODO())
	s.Require().NoError(err)
	s.Require().NotNil(list)

	s.Assert().True(slices.ContainsFunc(list, func(elem repositories.JetstreamBinding) bool {
		return elem.ID.String() == jb.ID.String()
	}))

	for _, elem := range list {
		if elem.ID == jb.ID {
			s.Assert().Equal(jb.ID, elem.ID)
			s.Assert().Equal(jb.LambdaARN, elem.LambdaARN)
			s.Assert().Equal(jb.Stream, elem.Stream)
			s.Assert().Equal(jb.ID, elem.Consumer)
			s.Assert().Equal(jb.Subject, elem.Subject)
			s.Assert().Equal(jb.MaxMessages, elem.MaxMessages)
			s.Assert().Equal(jb.MaxLatency, elem.MaxLatency)
			s.Assert().Equal(jb.DeliveryPolicy, elem.DeliveryPolicy)
			s.Assert().Equal(jb.AssignedPeerID, elem.AssignedPeerID)
		}
	}
}

func (s *BindingsConformanceSuite) TestDeleteJetstreamBinding() {
	jb, err := s.Candidate.CreateJetstreamBinding(context.TODO(), &repositories.CreateJetstreamBinding{
		LambdaARN:      "arn:aws:lambda:us-east-1:123456789012:function:my-function",
		Stream:         "my-stream",
		Subject:        "my-subject",
		MaxMessages:    10,
		MaxLatency:     5 * time.Second,
		DeliveryPolicy: "all",
	})
	s.Require().NoError(err)
	s.Require().NotNil(jb)

	err = s.Candidate.DeleteJetstreamBinding(context.TODO(), jb.ID)
	s.Require().NoError(err)

	list, err := s.Candidate.ListJetstreamBindings(context.TODO())
	s.Require().NoError(err)
	s.Require().NotNil(list)

	s.Assert().False(slices.ContainsFunc(list, func(elem repositories.JetstreamBinding) bool {
		return elem.ID == jb.ID
	}))
}

func NewBindingsConformanceSuite(peers repositories.Peers, candidate repositories.Bindings) *BindingsConformanceSuite {
	return &BindingsConformanceSuite{
		Suite:     &suite.Suite{},
		Peers:     peers,
		Candidate: candidate,
	}
}
