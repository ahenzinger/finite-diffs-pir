from costs_asymptotic import *
import matplotlib.pyplot as plt
from argparse import ArgumentParser
from multicolorline import colored_line

### utils for generating plots
ymax = 5

# remove values that are subsumed by other values + rebalancing
def prune_worse_values(xs, ys):
    # abs(slope) = (y-1)/(1-x)
    cur_slope = float('inf')
    ret_xs = []
    ret_ys = []
    for x, y in zip(xs, ys):
        tmp_slope = (y-1)/(1-x)
        if tmp_slope < cur_slope:
            ret_xs.append(x)
            ret_ys.append(y)
            cur_slope = tmp_slope
        else:
            break
    ret_xs.append(1)
    ret_ys.append(1)
    return ret_xs, ret_ys

def format_plot(s, ax, fdiffs=None, arrow_xy_override=None, fontsize=None, show_xlabel=True):
    # assert show_lbls in ['none', 'label', 'legend']
    global ymax
    ax.axhline(y = 1, color='0.5', linestyle=':')

    if fdiffs is None:
        title = f'{s} servers'
    elif fdiffs == 1:
        title = f'{s} servers, with finite differences'
    else:
        title = f'{s} servers, without finite differences'

    ax.set_title(title, fontsize=fontsize)
    if show_xlabel:
        ax.set_xlabel('Exponent in server online time', fontsize=fontsize)
    ax.set_ylabel('Exponent in preprocessing time', fontsize=fontsize)
    ax.set_xlim(1/s, 1)
    ax.set_ylim(0.8, ymax)

    if arrow_xy_override is None:
        arrow_xy_override = (0.7, 0.7)
    x, y = arrow_xy_override
    ax.annotate("better", xy=(x-0.1, y-0.1), xycoords=ax.transAxes,
        xytext=(x, y), textcoords='axes fraction',
        arrowprops=dict(arrowstyle='-|>', color='black'))

def plot_our_schemes(s, q, fig, ax, show_lbls='none', color=False, fdiffs=None, fontsize=None):
    assert show_lbls in ['none', 'label']
    global ymax

    s2degs = {
        2: [1],
        3: [1, 2],
        5: [1, 2, 4],
        11: [1, 4, 7, 10],
    }

    s2offsets = {
        2: (0.05, 0.05),
        3: (0.05, 0.05),
        5: (0.075, 0),
        11: (0.08, 0.1),
    }

    # our schemes
    individual_degs = s2degs[s] if color else sorted(list(set([1, q-1])))
    fdiffs_list = [False, True] if fdiffs is None else [fdiffs]
    cbar_done = False
    for d in individual_degs:
        for use_finite_differences in fdiffs_list:
            if d == 1 and (not use_finite_differences): # this is GLMDS theorem 6.4, will plot this separately
                continue
            xs = []
            ys = []
            theta = 0.01
            thetas = []
            while theta <= d/2:
                xs.append(server_online_time(theta, d, s, q, use_finite_differences=use_finite_differences))
                ys.append(preprocessing_exponent(theta, d, s, q, use_finite_differences=use_finite_differences))
                thetas.append(theta/d)
                theta += 0.01

            if color:
                lines = colored_line(
                    xs,
                    ys,
                    thetas,
                    ax,
                    linewidth=3,
                    cmap="plasma",
                )

                midx, midy = xs[int(len(xs)/2)], ys[int(len(ys)/2)]
                offsetx, offsety = s2offsets[s]
                ax.text(midx-offsetx, midy-offsety, f'd={d}')

                if (show_lbls != 'none') and (not cbar_done):
                    cbar = fig.colorbar(lines)
                    cbar.set_label('theta' if s == 2 else 'theta/d')
                    cbar_done = True
            else:
                if not use_finite_differences:
                    assert d != 1
                    lbl = 'This work (individual degree > 1, precomputed derivatives)'
                else:
                    if d == 1:
                        lbl = 'This work (individual degree 1, finite differences)' if s > 2 else 'This work (finite differences)'
                    else:
                        lbl = 'This work (individual degree > 1, finite differences)'

                xs, ys = prune_worse_values(xs, ys)

                ax.plot(
                    xs,
                    ys,
                    color='red' if use_finite_differences else 'orange',
                    linestyle='-' if d == 1 else '--',
                    label = lbl if show_lbls != 'none' else None,
                )

    # special cases
    if s == 2 and q == 2:
        for idx, theta in enumerate([0.2, 0.5]):
            x = H(theta/2)/H(theta)
            y = 1/H(theta)
            ax.scatter(
                x,
                y,
                marker='*',
                s=100,
                color='red',
                label="This work, corollary 3.2" if idx == 0 else None,
            )
    elif fdiffs != 0:
        x = server_online_time((q-1)/2, q-1, s, q, use_finite_differences=True)
        y = preprocessing_exponent((q-1)/2, q-1, s, q, use_finite_differences=True)
        ax.scatter(
            x,
            y,
            marker='*',
            s=100,
            color='red',
            label="This work, corollary 4.4" if (show_lbls != 'none') else None,
        )

def plot_glmds_schemes(s, q, ax, convex=True, lbl_6_4=None, lbl_4_9=None):
    # GLMDS24, theorem 6.4
    xs = []
    ys = []
    theta = 0.01
    while theta <= 0.5:
        x, y = glmds_theorem_6_4(theta, s, q)
        xs.append(x)
        ys.append(y)
        theta += 0.01
    if convex:
        xs, ys = prune_worse_values(xs, ys)
        ax.fill_between(
            xs,
            ys,
            ymax,
            color='0.8',
        )
    
    ax.plot(
        xs,
        ys,
        color='black',
        linestyle='-',
        label=lbl_6_4, # "GLM+24, theorem 6.4 (individual degree 1, precomputed derivatives)"
    )

    if s >= 4:
        # GLMDS24, theorem 4.9. resulting scheme is trivial for s = 2, 3
        x, y = glmds_theorem_4_9(s, q)
        ax.scatter(
            x,
            y,
            marker='s',
            color='black',
            label=lbl_4_9, # 'GLM+24, theorem 4.9 (only restrict individual degree)'
        )

if __name__ == '__main__':
    parser = ArgumentParser(
        description='generate comparison plots for DEPIR'
    )

    parser.add_argument(
        '--two-server-path', type=str, default=None
    )

    parser.add_argument(
        '--multi-server-path', type=str, default=None
    )

    parser.add_argument(
        '--detailed-two-server-path', type=str, default=None,
    )

    parser.add_argument(
        '--detailed-multi-server-path', type=str, default=None,
    )

    args = parser.parse_args()

    if args.two_server_path is not None:
        # 2 server case
        fig, ax = plt.subplots(nrows=1, ncols=1, figsize=(7, 5))
        plot_our_schemes(2, 2, fig, ax, show_lbls='label')
        plot_glmds_schemes(2, 2, ax,
            lbl_6_4="BIM00, theorem 4 and GLM+24, theorem 6.4 (precomputed derivatives)",
        )
        format_plot(2, ax)
        fig.legend(bbox_to_anchor=(0.9, 0.88), fontsize='small')
        fig.savefig(args.two_server_path)
        del fig, ax

    if args.multi_server_path is not None:
        # multi-server case
        fig, ax = plt.subplots(nrows=1, ncols=4, figsize=(32, 5))
        for idx, q in enumerate([3, 5, 11]):
            plot_our_schemes(q, q, fig, ax[idx], show_lbls='label' if q == 11 else 'none')
            plot_glmds_schemes(q, q, ax[idx],
                lbl_6_4="GLM+24, theorem 6.4 (individual degree 1, precomputed derivatives)" if q == 11 else None,
                lbl_4_9='GLM+24, theorem 4.9 (only restrict individual degree)' if q == 11 else None,
            )
            format_plot(q, ax[idx],
                arrow_xy_override=(0.8, 0.8),
                fontsize='xx-large',
            )
        ax[3].set_axis_off() # the 4th subplot is just a hack to fit the legend in
        fig.legend(loc='center right', fontsize='xx-large')
        fig.savefig(args.multi_server_path)
        del fig, ax

    if args.detailed_two_server_path is not None:
        fig, ax = plt.subplots(nrows=1, ncols=1, figsize=(7, 5))
        plot_glmds_schemes(2, 2, ax,
            lbl_6_4="BIM00 and GLM+24 (precomputed derivatives)",
            convex=False,
        )
        plot_our_schemes(2, 2, fig, ax, show_lbls='label', color=True)
        format_plot(2, ax)
        fig.legend(bbox_to_anchor=(0.73, 0.88), fontsize='small')
        fig.savefig(args.detailed_two_server_path)
        del fig, ax

    if args.detailed_multi_server_path is not None:
        fig, ax = plt.subplots(nrows=2, ncols=3, figsize=(25, 10))
        for fdiffs in [0, 1]:
            for idx, q in enumerate([3, 5, 11]):
                if fdiffs == 0:
                    plot_glmds_schemes(q, q, ax[fdiffs, idx],
                        lbl_6_4="GLM+24, theorem 6.4 (d = 1)" if q == 11 else None,
                        lbl_4_9="GLM+24, theorem 4.9 (theta = d)" if q == 11 else None,
                        convex=False,
                    )
                plot_our_schemes(q, q, fig, ax[fdiffs, idx], show_lbls='label' if q == 11 else 'none', color=True, fdiffs=fdiffs, fontsize='xx-large')
                format_plot(q, ax[fdiffs, idx],
                    fdiffs=fdiffs,
                    fontsize='xx-large',
                    show_xlabel=(fdiffs==1),
                )

        fig.legend(loc='upper center', fontsize='xx-large', ncols=4)
        fig.savefig(args.detailed_multi_server_path)