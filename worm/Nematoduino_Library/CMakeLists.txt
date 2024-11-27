add_subdirectory(utility)

file(GLOB SOURCES "*.cpp" "*.c")
file(GLOB HEADERS "*.hpp" "*.h")

add_library(nematoduino ${SOURCES} ${HEADERS})

target_include_directories(nematoduino PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

target_link_libraries(nematoduino PRIVATE utility)
