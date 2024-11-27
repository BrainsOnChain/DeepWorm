use std::ffi::{c_int, c_void};
use std::thread::sleep;
use std::time::Duration;

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
    let worm = unsafe { Worm_Worm() };

    let mut pos_x = 0.0;
    let mut pos_y = 0.0;

    let mut direction = 0.0;

    loop {
        unsafe { Worm_chemotaxis(worm) };

        let left = unsafe { Worm_getLeftMuscle(worm) };
        let right = unsafe { Worm_getRightMuscle(worm) };

        println!("Left: {}, Right: {}", left, right);

        direction += -(right - left) as f64 / 5.0;
        direction = direction.rem_euclid(360.0);
        let distance = (right + left) as f64 / 100.0;

        let new_pos_x = pos_x + (direction * 3.14 / 180.0).sin() * distance;
        let new_pos_y = pos_y + (direction * 3.14 / 180.0).cos() * distance;

        pos_x = new_pos_x;
        pos_y = new_pos_y;

        println!("D: {}, X: {}, Y: {}", direction, pos_x, pos_y);

        sleep(Duration::from_millis(16));
    }
}
