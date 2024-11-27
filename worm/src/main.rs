use std::ffi::{c_int, c_void};

#[link(name = "nematoduino")]
extern "C" {
    fn Worm_Worm() -> *mut c_void;
    fn Worm_destroy(worm: *mut c_void);
    fn Worm_chemotaxis(worm: *mut c_void);
    fn Worm_noseTouch(worm: *mut c_void);
    fn Worm_getLeftMuscle(worm: *mut c_void) -> c_int;
    fn Worm_getRightMuscle(worm: *mut c_void) -> c_int;
}

fn main() {
    println!("Hello, world!");
}
