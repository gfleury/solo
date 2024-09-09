export function OneOrBothBySize(big: any, small: any) {
  if (window.innerWidth > 800)
    return (
      <>
        {small} {big}
      </>
    );
  return small;
}

export function OneOrOtherBySize(big: any, small: any) {
  if (window.innerWidth > 800) return big;
  return small;
}
