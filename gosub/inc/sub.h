#ifndef SUB_H
#define SUB_H

#include "ros/ros.h"
#include "std_msgs/String.h"

class sub{
public:
	sub(){};
	void Init(int argc, char **argv);
	//void Reciever(const std_msgs::String::ConstPtr& msg);
};

#endif