#include "sub.h"
#include "ros/ros.h"
#include "std_msgs/String.h"

void chatterCallback(const std_msgs::String::ConstPtr& msg){
	ROS_INFO("I heard: [%s]", msg->data.c_str());
}

void sub::Init(int argc, char **argv){
	// Initial
	ros::init(argc, argv, "sub");

	// Create a ros::Subscriber object
	ros::NodeHandle n;

	// Subscribe to topic, buffer, callback function
	ros::Subscriber sub = n.subscribe("/vicon/Jet/Jet", 1000, chatterCallback);

	// Listen
	ros::spin();
};