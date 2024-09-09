/**
 *
 * @export
 * @interface Node
 */
export interface NetworkNode {
  ID?: number;
  PeerID?: string;
  Hostname?: string;
  OS?: string;
  Arch?: string;
  IP?: string;
  Version?: string;
  LocalRoutes?: Array<string>;
}

/**
 *
 * @export
 * @interface Network
 */
export interface Network {
  ID?: number;
  name: string;
  cidr: string;
  connection_config: string;
  nodes?: Array<NetworkNode>;
  user: User;
  linkedusers: Array<LinkedUser>;
}

/**
 *
 * @export
 * @interface LinkedUser
 */
export interface LinkedUser {
  user?: User;
  permissions?: LinkedUser.PermissionsEnum;
}

/**
 * @export
 * @namespace LinkedUser
 */
export namespace LinkedUser {
  export enum PermissionsEnum {
    Write = "write",
    Read = "read",
    Owner = "owner",
  }
}

/**
 *
 * @export
 * @interface User
 */
export interface User {
  ID?: number;
  username?: string;
  email?: string;
  userStatus?: number;
  networks?: Array<Network>;
  avatar?: string;
}
