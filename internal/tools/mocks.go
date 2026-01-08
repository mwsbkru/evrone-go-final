package tools

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
)

// MockClusterAdmin is a mock implementation of sarama.ClusterAdmin
type MockClusterAdmin struct {
	mock.Mock
}

func (m *MockClusterAdmin) CreateTopic(topic string, detail *sarama.TopicDetail, validateOnly bool) error {
	args := m.Called(topic, detail, validateOnly)
	return args.Error(0)
}

func (m *MockClusterAdmin) DeleteTopic(topic string) error {
	args := m.Called(topic)
	return args.Error(0)
}

func (m *MockClusterAdmin) CreatePartitions(topic string, count int32, assignment [][]int32, validateOnly bool) error {
	args := m.Called(topic, count, assignment, validateOnly)
	return args.Error(0)
}

func (m *MockClusterAdmin) AlterPartitionReassignments(topic string, assignment [][]int32) error {
	args := m.Called(topic, assignment)
	return args.Error(0)
}

func (m *MockClusterAdmin) ListPartitionReassignments(topics string, partitions []int32) (map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus, error) {
	args := m.Called(topics, partitions)
	return args.Get(0).(map[string]map[int32]*sarama.PartitionReplicaReassignmentsStatus), args.Error(1)
}

func (m *MockClusterAdmin) AlterClientQuotas(entity []sarama.QuotaEntityComponent, op sarama.ClientQuotasOp, validateOnly bool) error {
	args := m.Called(entity, op, validateOnly)
	return args.Error(0)
}

func (m *MockClusterAdmin) Controller() (*sarama.Broker, error) {
	args := m.Called()
	return args.Get(0).(*sarama.Broker), args.Error(1)
}

func (m *MockClusterAdmin) Coordinator(group string) (*sarama.Broker, error) {
	args := m.Called(group)
	return args.Get(0).(*sarama.Broker), args.Error(1)
}

func (m *MockClusterAdmin) IncrementalAlterConfig(resourceType sarama.ConfigResourceType, name string, entries map[string]sarama.IncrementalAlterConfigsEntry, validateOnly bool) error {
	args := m.Called(resourceType, name, entries, validateOnly)
	return args.Error(0)
}

func (m *MockClusterAdmin) CreateACLs(resources []*sarama.ResourceAcls) error {
	args := m.Called(resources)
	return args.Error(0)
}

func (m *MockClusterAdmin) ElectLeaders(electionType sarama.ElectionType, topicPartitions map[string][]int32) (map[string]map[int32]*sarama.PartitionResult, error) {
	args := m.Called(electionType, topicPartitions)
	return args.Get(0).(map[string]map[int32]*sarama.PartitionResult), args.Error(1)
}

func (m *MockClusterAdmin) DeleteConsumerGroupOffset(group string, topic string, partition int32) error {
	args := m.Called(group, topic, partition)
	return args.Error(0)
}

func (m *MockClusterAdmin) DescribeUserScramCredentials(users []string) ([]*sarama.DescribeUserScramCredentialsResult, error) {
	args := m.Called(users)
	return args.Get(0).([]*sarama.DescribeUserScramCredentialsResult), args.Error(1)
}

func (m *MockClusterAdmin) DeleteUserScramCredentials(delete []sarama.AlterUserScramCredentialsDelete) ([]*sarama.AlterUserScramCredentialsResult, error) {
	args := m.Called(delete)
	return args.Get(0).([]*sarama.AlterUserScramCredentialsResult), args.Error(1)
}

func (m *MockClusterAdmin) UpsertUserScramCredentials(upsert []sarama.AlterUserScramCredentialsUpsert) ([]*sarama.AlterUserScramCredentialsResult, error) {
	args := m.Called(upsert)
	return args.Get(0).([]*sarama.AlterUserScramCredentialsResult), args.Error(1)
}

func (m *MockClusterAdmin) DescribeClientQuotas(components []sarama.QuotaFilterComponent, strict bool) ([]sarama.DescribeClientQuotasEntry, error) {
	args := m.Called(components, strict)
	return args.Get(0).([]sarama.DescribeClientQuotasEntry), args.Error(1)
}

func (m *MockClusterAdmin) DeleteRecords(topic string, partitionOffsets map[int32]int64) error {
	args := m.Called(topic, partitionOffsets)
	return args.Error(0)
}

func (m *MockClusterAdmin) DescribeConfig(resource sarama.ConfigResource) ([]sarama.ConfigEntry, error) {
	args := m.Called(resource)
	return args.Get(0).([]sarama.ConfigEntry), args.Error(1)
}

func (m *MockClusterAdmin) AlterConfig(resourceType sarama.ConfigResourceType, name string, entries map[string]*string, validateOnly bool) error {
	args := m.Called(resourceType, name, entries, validateOnly)
	return args.Error(0)
}

func (m *MockClusterAdmin) CreateACL(resource sarama.Resource, acl sarama.Acl) error {
	args := m.Called(resource, acl)
	return args.Error(0)
}

func (m *MockClusterAdmin) ListAcls(filter sarama.AclFilter) ([]sarama.ResourceAcls, error) {
	args := m.Called(filter)
	return args.Get(0).([]sarama.ResourceAcls), args.Error(1)
}

func (m *MockClusterAdmin) DeleteACL(filter sarama.AclFilter, validateOnly bool) ([]sarama.MatchingAcl, error) {
	args := m.Called(filter, validateOnly)
	return args.Get(0).([]sarama.MatchingAcl), args.Error(1)
}

func (m *MockClusterAdmin) ListConsumerGroups() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockClusterAdmin) DescribeConsumerGroups(groups []string) ([]*sarama.GroupDescription, error) {
	args := m.Called(groups)
	return args.Get(0).([]*sarama.GroupDescription), args.Error(1)
}

func (m *MockClusterAdmin) ListConsumerGroupOffsets(group string, topicPartitions map[string][]int32) (*sarama.OffsetFetchResponse, error) {
	args := m.Called(group, topicPartitions)
	return args.Get(0).(*sarama.OffsetFetchResponse), args.Error(1)
}

func (m *MockClusterAdmin) DeleteConsumerGroup(group string) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockClusterAdmin) DescribeCluster() (brokers []*sarama.Broker, controllerID int32, err error) {
	args := m.Called()
	return args.Get(0).([]*sarama.Broker), args.Get(1).(int32), args.Error(2)
}

func (m *MockClusterAdmin) DescribeLogDirs(brokers []int32) (map[int32][]sarama.DescribeLogDirsResponseDirMetadata, error) {
	args := m.Called(brokers)
	return args.Get(0).(map[int32][]sarama.DescribeLogDirsResponseDirMetadata), args.Error(1)
}

func (m *MockClusterAdmin) RemoveMemberFromConsumerGroup(groupID string, groupInstanceIds []string) (*sarama.LeaveGroupResponse, error) {
	args := m.Called(groupID, groupInstanceIds)
	return args.Get(0).(*sarama.LeaveGroupResponse), args.Error(1)
}

func (m *MockClusterAdmin) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClusterAdmin) ListTopics() (map[string]sarama.TopicDetail, error) {
	args := m.Called()
	return args.Get(0).(map[string]sarama.TopicDetail), args.Error(1)
}

func (m *MockClusterAdmin) DescribeTopics(topics []string) (metadata []*sarama.TopicMetadata, err error) {
	args := m.Called(topics)
	return args.Get(0).([]*sarama.TopicMetadata), args.Error(1)
}

func (m *MockClusterAdmin) DescribeTopicPartitions(topics []string) (map[string][]sarama.PartitionMetadata, error) {
	args := m.Called(topics)
	return args.Get(0).(map[string][]sarama.PartitionMetadata), args.Error(1)
}
