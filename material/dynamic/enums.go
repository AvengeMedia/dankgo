package dynamic

// Variant indicates type of scheme to generate
//
// ENUM(monochrome, neutral, tonal_spot, vibrant, expressive, fidelity, content,
// rainbow, fruit_salad)
type Variant uint

// Version indicates the material color specification year
//
// ENUM(2021=2021, 2025=2025)
type Version uint

// Platform indicates target platform for generating colors
//
// ENUM(phone, watch)
type Platform uint

// TonePolarity describes the difference in tone between colors.
//
// ENUM(darker, lighter, nearer, farther, relative_darker, relative_lighter)
type TonePolarity uint

// Constraint describes how to fulfill a tone delta pair constraint.
//
// ENUM(exact, nearer, farther)
type Constraint uint
