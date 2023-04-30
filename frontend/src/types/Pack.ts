export interface PackageMetadata {
  Name: string;
  Version?: string;
  ID: string;
}

export interface PackageData {
  Content?: string;
  URL?: string;
  JSProgram?: string;
}

export interface Package {
  metadata: PackageMetadata;
  data: PackageData;
}

export interface PackageRating {
  BusFactor: number;
  Correctness: number;
  RampUp: number;
  ResponsiveMaintainer: number;
  LicenseScore: number;
  GoodPinningPractice: number;
  PullRequest: number;
  NetScore: number;
}

export interface PackageQuery {
  Version?: string;
  Name: string;
}

export interface PackageRegEx {
  RegEx: string;
}