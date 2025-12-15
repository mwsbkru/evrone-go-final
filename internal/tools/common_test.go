package tools

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestTopicExists_Error(t *testing.T) {
	mockAdmin := new(MockClusterAdmin)
	expectedErr := errors.New("failed to list topics")

	mockAdmin.On("ListTopics").Return(map[string]sarama.TopicDetail(nil), expectedErr)

	exists, err := topicExists(mockAdmin, "test-topic")
	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "can`t fetch list of topics")
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockAdmin.AssertExpectations(t)
}

// TestTopicExists_MultipleTopics tests with multiple topics in the map
func TestTopicExists_MultipleTopics(t *testing.T) {
	mockAdmin := new(MockClusterAdmin)

	topics := map[string]sarama.TopicDetail{
		"topic1": {NumPartitions: 1, ReplicationFactor: 1},
		"topic2": {NumPartitions: 2, ReplicationFactor: 1},
		"topic3": {NumPartitions: 3, ReplicationFactor: 1},
	}

	mockAdmin.On("ListTopics").Return(topics, nil)

	// Test existing topic
	exists, err := topicExists(mockAdmin, "topic2")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing topic
	exists, err = topicExists(mockAdmin, "topic4")
	assert.NoError(t, err)
	assert.False(t, exists)

	mockAdmin.AssertExpectations(t)
}

func TestCreateTopic_Success(t *testing.T) {
	mockAdmin := new(MockClusterAdmin)
	topicName := "test-topic"
	partitions := int32(3)
	replicationFactor := int16(1)

	expectedTopicDetail := sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	mockAdmin.On("CreateTopic", topicName, &expectedTopicDetail, false).Return(nil)

	err := createTopic(mockAdmin, topicName, partitions, replicationFactor)
	assert.NoError(t, err)

	mockAdmin.AssertExpectations(t)
}

func TestCreateTopic_Error(t *testing.T) {
	mockAdmin := new(MockClusterAdmin)
	topicName := "test-topic"
	partitions := int32(3)
	replicationFactor := int16(1)
	expectedErr := errors.New("topic creation failed")

	expectedTopicDetail := sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	mockAdmin.On("CreateTopic", topicName, &expectedTopicDetail, false).Return(expectedErr)

	err := createTopic(mockAdmin, topicName, partitions, replicationFactor)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cant create topic")
	assert.Contains(t, err.Error(), expectedErr.Error())

	mockAdmin.AssertExpectations(t)
}
