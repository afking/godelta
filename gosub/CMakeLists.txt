cmake_minimum_required(VERSION 2.8 FATAL_ERROR)
# Above line indicates to Cmake that minimum version 2.8 is required.
# As you may have noted all lines beginning with hash symbol are treated as comments by Cmake.
SET(sub)
# In the above line we declared a variable sampleName and assigned it a value MyApp. We will make use of the same later.project(PrjMyApp)
# here we define the project or solution (for visual studio) name


# In following lines we try to find out the packages of additional libraries, if reqd. Here we are trying to locate PCL and mrpt library.
# u may replace the names by yours. Each such package should have a Cmake file in them, if not then we may have to manually define the paths.
# we Show that later.
find_package(catkin REQUIRED COMPONENTS
	geometry_msgs
	rospy
	std_msgs
	message_generation
)

## Generate added messages and services with any dependencies listed here
generate_messages(
	DEPENDENCIES
	geometry_msgs
	std_msgs
)

catkin_package(
	# INCLUDE_DIRS include
	# LIBRARIES crazyflie
	CATKIN_DEPENDS message_runtime geometry_msgs rospy std_msgs
	# DEPENDS system_lib
)
###########
## Build ##
###########
## Specify additional locations of header files
## Your package locations should be listed before other locations
# include_directories(include)
include_directories(
	${catkin_INCLUDE_DIRS}
)