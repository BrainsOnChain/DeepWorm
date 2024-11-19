add_subdirectory(utility)

file(GLOB SOURCES "*.cpp")
file(GLOB HEADERS "*.hpp")

add_library(nematoduino ${SOURCES} ${HEADERS})

target_include_directories(nematoduino PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(nematoduino PRIVATE utility)
