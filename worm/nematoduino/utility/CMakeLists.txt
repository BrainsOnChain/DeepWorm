file(GLOB SOURCES "*.c")
file(GLOB HEADERS "*.h")

add_library(utility ${SOURCES} ${HEADERS})

target_include_directories(utility PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

install(TARGETS utility
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)
